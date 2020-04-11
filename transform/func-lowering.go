package transform

// This file lowers func values into their final form. This is necessary for
// funcValueSwitch, which needs full program analysis.

import (
	"sort"
	"strconv"

	"github.com/tinygo-org/tinygo/compiler/llvmutil"
	"tinygo.org/x/go-llvm"
)

// funcSignatureInfo keeps information about a single signature and its uses.
type funcSignatureInfo struct {
	sig                     llvm.Value   // *uint8 to identify the signature
	funcValueWithSignatures []llvm.Value // slice of runtime.funcValueWithSignature
}

// funcWithUses keeps information about a single function used as func value and
// the assigned function ID. More commonly used functions are assigned a lower
// ID.
type funcWithUses struct {
	funcPtr  llvm.Value
	useCount int // how often this function is used in a func value
	id       int // assigned ID
}

// Slice to sort functions by their use counts, or else their name if they're
// used equally often.
type funcWithUsesList []*funcWithUses

func (l funcWithUsesList) Len() int { return len(l) }
func (l funcWithUsesList) Less(i, j int) bool {
	if l[i].useCount != l[j].useCount {
		// return the reverse: we want the highest use counts sorted first
		return l[i].useCount > l[j].useCount
	}
	iName := l[i].funcPtr.Name()
	jName := l[j].funcPtr.Name()
	return iName < jName
}
func (l funcWithUsesList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

// LowerFuncValues lowers the runtime.funcValueWithSignature type and
// runtime.getFuncPtr function to their final form.
func LowerFuncValues(mod llvm.Module) {
	ctx := mod.Context()
	builder := ctx.NewBuilder()
	uintptrType := ctx.IntType(llvm.NewTargetData(mod.DataLayout()).PointerSize() * 8)

	// Find all func values used in the program with their signatures.
	funcValueWithSignaturePtr := llvm.PointerType(mod.GetTypeByName("runtime.funcValueWithSignature"), 0)
	signatures := map[string]*funcSignatureInfo{}
	for global := mod.FirstGlobal(); !global.IsNil(); global = llvm.NextGlobal(global) {
		if global.Type() != funcValueWithSignaturePtr {
			continue
		}
		sig := llvm.ConstExtractValue(global.Initializer(), []uint32{1})
		name := sig.Name()
		if info, ok := signatures[name]; ok {
			info.funcValueWithSignatures = append(info.funcValueWithSignatures, global)
		} else {
			signatures[name] = &funcSignatureInfo{
				sig:                     sig,
				funcValueWithSignatures: []llvm.Value{global},
			}
		}
	}

	// Sort the signatures, for deterministic execution.
	names := make([]string, 0, len(signatures))
	for name := range signatures {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		info := signatures[name]
		functions := make(funcWithUsesList, len(info.funcValueWithSignatures))
		for i, use := range info.funcValueWithSignatures {
			var useCount int
			for _, use2 := range getUses(use) {
				useCount += len(getUses(use2))
			}
			functions[i] = &funcWithUses{
				funcPtr:  llvm.ConstExtractValue(use.Initializer(), []uint32{0}).Operand(0),
				useCount: useCount,
			}
		}
		sort.Sort(functions)

		for i, fn := range functions {
			fn.id = i + 1
			for _, ptrtoint := range getUses(fn.funcPtr) {
				if ptrtoint.IsAConstantExpr().IsNil() || ptrtoint.Opcode() != llvm.PtrToInt {
					continue
				}
				for _, funcValueWithSignatureConstant := range getUses(ptrtoint) {
					if !funcValueWithSignatureConstant.IsACallInst().IsNil() && funcValueWithSignatureConstant.CalledValue().Name() == "runtime.makeGoroutine" {
						// makeGoroutine calls are handled seperately
						continue
					}
					for _, funcValueWithSignatureGlobal := range getUses(funcValueWithSignatureConstant) {
						for _, use := range getUses(funcValueWithSignatureGlobal) {
							if ptrtoint.IsAConstantExpr().IsNil() || ptrtoint.Opcode() != llvm.PtrToInt {
								panic("expected const ptrtoint")
							}
							use.ReplaceAllUsesWith(llvm.ConstInt(uintptrType, uint64(fn.id), false))
						}
					}
				}
			}
		}

		for _, getFuncPtrCall := range getUses(info.sig) {
			if getFuncPtrCall.IsACallInst().IsNil() {
				continue
			}
			if getFuncPtrCall.CalledValue().Name() != "runtime.getFuncPtr" {
				panic("expected all call uses to be runtime.getFuncPtr")
			}
			funcID := getFuncPtrCall.Operand(1)
			switch len(functions) {
			case 0:
				// There are no functions used in a func value that implement
				// this signature. The only possible value is a nil value.
				for _, inttoptr := range getUses(getFuncPtrCall) {
					if inttoptr.IsAIntToPtrInst().IsNil() {
						panic("expected inttoptr")
					}
					nilptr := llvm.ConstPointerNull(inttoptr.Type())
					inttoptr.ReplaceAllUsesWith(nilptr)
					inttoptr.EraseFromParentAsInstruction()
				}
				getFuncPtrCall.EraseFromParentAsInstruction()
			case 1:
				// There is exactly one function with this signature that is
				// used in a func value. The func value itself can be either nil
				// or this one function.
				builder.SetInsertPointBefore(getFuncPtrCall)
				zero := llvm.ConstInt(uintptrType, 0, false)
				isnil := builder.CreateICmp(llvm.IntEQ, funcID, zero, "")
				funcPtrNil := llvm.ConstPointerNull(functions[0].funcPtr.Type())
				funcPtr := builder.CreateSelect(isnil, funcPtrNil, functions[0].funcPtr, "")
				for _, inttoptr := range getUses(getFuncPtrCall) {
					if inttoptr.IsAIntToPtrInst().IsNil() {
						panic("expected inttoptr")
					}
					inttoptr.ReplaceAllUsesWith(funcPtr)
					inttoptr.EraseFromParentAsInstruction()
				}
				getFuncPtrCall.EraseFromParentAsInstruction()
			default:
				// There are multiple functions used in a func value that
				// implement this signature.
				// What we'll do is transform the following:
				//     rawPtr := runtime.getFuncPtr(func.ptr)
				//     if rawPtr == nil {
				//         runtime.nilPanic()
				//     }
				//     result := rawPtr(...args, func.context)
				// into this:
				//     if false {
				//         runtime.nilPanic()
				//     }
				//     var result // Phi
				//     switch fn.id {
				//     case 0:
				//         runtime.nilPanic()
				//     case 1:
				//         result = call first implementation...
				//     case 2:
				//         result = call second implementation...
				//     default:
				//         unreachable
				//     }

				// Remove some casts, checks, and the old call which we're going
				// to replace.
				for _, callIntPtr := range getUses(getFuncPtrCall) {
					if !callIntPtr.IsACallInst().IsNil() && callIntPtr.CalledValue().Name() == "internal/task.start" {
						// Special case for goroutine starts.
						addFuncLoweringSwitch(mod, builder, funcID, callIntPtr, func(funcPtr llvm.Value, params []llvm.Value) llvm.Value {
							i8ptrType := llvm.PointerType(ctx.Int8Type(), 0)
							calleeValue := builder.CreatePtrToInt(funcPtr, uintptrType, "")
							start := mod.NamedFunction("internal/task.start")
							builder.CreateCall(start, []llvm.Value{calleeValue, callIntPtr.Operand(1), llvm.Undef(i8ptrType), llvm.ConstNull(i8ptrType)}, "")
							return llvm.Value{} // void so no return value
						}, functions)
						callIntPtr.EraseFromParentAsInstruction()
						continue
					}
					if callIntPtr.IsAIntToPtrInst().IsNil() {
						panic("expected inttoptr")
					}
					for _, ptrUse := range getUses(callIntPtr) {
						if !ptrUse.IsAICmpInst().IsNil() {
							ptrUse.ReplaceAllUsesWith(llvm.ConstInt(ctx.Int1Type(), 0, false))
						} else if !ptrUse.IsACallInst().IsNil() && ptrUse.CalledValue() == callIntPtr {
							addFuncLoweringSwitch(mod, builder, funcID, ptrUse, func(funcPtr llvm.Value, params []llvm.Value) llvm.Value {
								return builder.CreateCall(funcPtr, params, "")
							}, functions)
						} else {
							panic("unexpected getFuncPtrCall")
						}
						ptrUse.EraseFromParentAsInstruction()
					}
					callIntPtr.EraseFromParentAsInstruction()
				}
				getFuncPtrCall.EraseFromParentAsInstruction()
			}
		}
	}
}

// addFuncLoweringSwitch creates a new switch on a function ID and inserts calls
// to the newly created direct calls. The funcID is the number to switch on,
// call is the call instruction to replace, and createCall is the callback that
// actually creates the new call. By changing createCall to something other than
// builder.CreateCall, instead of calling a function it can start a new
// goroutine for example.
func addFuncLoweringSwitch(mod llvm.Module, builder llvm.Builder, funcID, call llvm.Value, createCall func(funcPtr llvm.Value, params []llvm.Value) llvm.Value, functions funcWithUsesList) {
	ctx := mod.Context()
	uintptrType := ctx.IntType(llvm.NewTargetData(mod.DataLayout()).PointerSize() * 8)
	i8ptrType := llvm.PointerType(ctx.Int8Type(), 0)

	// The block that cannot be reached with correct funcValues (to help the
	// optimizer).
	builder.SetInsertPointBefore(call)
	defaultBlock := ctx.AddBasicBlock(call.InstructionParent().Parent(), "func.default")
	builder.SetInsertPointAtEnd(defaultBlock)
	builder.CreateUnreachable()

	// Create the switch.
	builder.SetInsertPointBefore(call)
	sw := builder.CreateSwitch(funcID, defaultBlock, len(functions)+1)

	// Split right after the switch. We will need to insert a few basic blocks
	// in this gap.
	nextBlock := llvmutil.SplitBasicBlock(builder, sw, llvm.NextBasicBlock(sw.InstructionParent()), "func.next")

	// The 0 case, which is actually a nil check.
	nilBlock := ctx.InsertBasicBlock(nextBlock, "func.nil")
	builder.SetInsertPointAtEnd(nilBlock)
	nilPanic := mod.NamedFunction("runtime.nilPanic")
	builder.CreateCall(nilPanic, []llvm.Value{llvm.Undef(i8ptrType), llvm.ConstNull(i8ptrType)}, "")
	builder.CreateUnreachable()
	sw.AddCase(llvm.ConstInt(uintptrType, 0, false), nilBlock)

	// Gather the list of parameters for every call we're going to make.
	callParams := make([]llvm.Value, call.OperandsCount()-1)
	for i := range callParams {
		callParams[i] = call.Operand(i)
	}

	// If the call produces a value, we need to get it using a PHI
	// node.
	phiBlocks := make([]llvm.BasicBlock, len(functions))
	phiValues := make([]llvm.Value, len(functions))
	for i, fn := range functions {
		// Insert a switch case.
		bb := ctx.InsertBasicBlock(nextBlock, "func.call"+strconv.Itoa(fn.id))
		builder.SetInsertPointAtEnd(bb)
		result := createCall(fn.funcPtr, callParams)
		builder.CreateBr(nextBlock)
		sw.AddCase(llvm.ConstInt(uintptrType, uint64(fn.id), false), bb)
		phiBlocks[i] = bb
		phiValues[i] = result
	}
	// Create the PHI node so that the call result flows into the
	// next block (after the split). This is only necessary when the
	// call produced a value.
	if call.Type().TypeKind() != llvm.VoidTypeKind {
		builder.SetInsertPointBefore(nextBlock.FirstInstruction())
		phi := builder.CreatePHI(call.Type(), "")
		phi.AddIncoming(phiValues, phiBlocks)
		call.ReplaceAllUsesWith(phi)
	}
}
