target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32-unknown-unknown-wasm"

%runtime.typecodeID = type { %runtime.typecodeID*, i32 }
%runtime.funcValueWithSignature = type { i32, %runtime.typecodeID* }

@"reflect/types.type:func:{basic:int8}{}" = external constant %runtime.typecodeID
@"reflect/types.type:func:{basic:uint8}{}" = external constant %runtime.typecodeID
@"reflect/types.type:func:{basic:int}{}" = external constant %runtime.typecodeID
@"funcInt8$withSignature" = constant %runtime.funcValueWithSignature { i32 ptrtoint (void (i8, i8*, i8*)* @funcInt8 to i32), %runtime.typecodeID* @"reflect/types.type:func:{basic:int8}{}" }
@"func1Uint8$withSignature" = constant %runtime.funcValueWithSignature { i32 ptrtoint (void (i8, i8*, i8*)* @func1Uint8 to i32), %runtime.typecodeID* @"reflect/types.type:func:{basic:uint8}{}" }
@"func2Uint8$withSignature" = constant %runtime.funcValueWithSignature { i32 ptrtoint (void (i8, i8*, i8*)* @func2Uint8 to i32), %runtime.typecodeID* @"reflect/types.type:func:{basic:uint8}{}" }
@"main$withSignature" = constant %runtime.funcValueWithSignature { i32 ptrtoint (void (i32, i8*, i8*)* @"main$1" to i32), %runtime.typecodeID* @"reflect/types.type:func:{basic:int}{}" }
@"main$2$withSignature" = constant %runtime.funcValueWithSignature { i32 ptrtoint (void (i32, i8*, i8*)* @"main$2" to i32), %runtime.typecodeID* @"reflect/types.type:func:{basic:int}{}" }

declare i32 @runtime.getFuncPtr(i8*, i32, %runtime.typecodeID*, i8*, i8*)

declare void @"internal/task.start"(i32, i8*, i8*, i8*)

declare void @runtime.nilPanic(i8*, i8*)

declare void @"main$1"(i32, i8*, i8*)

declare void @"main$2"(i32, i8*, i8*)

declare void @funcInt8(i8, i8*, i8*)

declare void @func1Uint8(i8, i8*, i8*)

declare void @func2Uint8(i8, i8*, i8*)

; Call a function of which only one function with this signature is used as a
; function value. This means that lowering it to IR is trivial: simply check
; whether the func value is nil, and if not, call that one function directly.
define void @runFunc1(i8*, i32, i8, i8* %context, i8* %parentHandle) {
entry:
  %3 = call i32 @runtime.getFuncPtr(i8* %0, i32 %1, %runtime.typecodeID* @"reflect/types.type:func:{basic:int8}{}", i8* undef, i8* null)
  %4 = inttoptr i32 %3 to void (i8, i8*, i8*)*
  %5 = icmp eq void (i8, i8*, i8*)* %4, null
  br i1 %5, label %fpcall.nil, label %fpcall.next

fpcall.nil:
  call void @runtime.nilPanic(i8* undef, i8* null)
  unreachable

fpcall.next:
  call void %4(i8 %2, i8* %0, i8* undef)
  ret void
}

; There are two functions with this signature used in a func value. That means
; that we'll have to check at runtime which of the two it is (or whether the
; func value is nil). This call will thus be lowered to a switch statement.
define void @runFunc2(i8*, i32, i8, i8* %context, i8* %parentHandle) {
entry:
  %3 = call i32 @runtime.getFuncPtr(i8* %0, i32 %1, %runtime.typecodeID* @"reflect/types.type:func:{basic:uint8}{}", i8* undef, i8* null)
  %4 = inttoptr i32 %3 to void (i8, i8*, i8*)*
  %5 = icmp eq void (i8, i8*, i8*)* %4, null
  br i1 %5, label %fpcall.nil, label %fpcall.next

fpcall.nil:
  call void @runtime.nilPanic(i8* undef, i8* null)
  unreachable

fpcall.next:
  call void %4(i8 %2, i8* %0, i8* undef)
  ret void
}

; Special case for internal/task.start.
define void @sleepFuncValue(i8*, i32, i8* nocapture readnone %context, i8* nocapture readnone %parentHandle) {
entry:
  %2 = call i32 @runtime.getFuncPtr(i8* %0, i32 %1, %runtime.typecodeID* @"reflect/types.type:func:{basic:int}{}", i8* undef, i8* null)
  call void @"internal/task.start"(i32 %2, i8* null, i8* undef, i8* null)
  ret void
}
