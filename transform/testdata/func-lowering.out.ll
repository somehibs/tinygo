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

define void @runFunc1(i8* %0, i32 %1, i8 %2, i8* %context, i8* %parentHandle) {
entry:
  %3 = icmp eq i32 %1, 0
  %4 = select i1 %3, void (i8, i8*, i8*)* null, void (i8, i8*, i8*)* @funcInt8
  %5 = icmp eq void (i8, i8*, i8*)* %4, null
  br i1 %5, label %fpcall.nil, label %fpcall.next

fpcall.nil:                                       ; preds = %entry
  call void @runtime.nilPanic(i8* undef, i8* null)
  unreachable

fpcall.next:                                      ; preds = %entry
  call void %4(i8 %2, i8* %0, i8* undef)
  ret void
}

define void @runFunc2(i8* %0, i32 %1, i8 %2, i8* %context, i8* %parentHandle) {
entry:
  br i1 false, label %fpcall.nil, label %fpcall.next

fpcall.nil:                                       ; preds = %entry
  call void @runtime.nilPanic(i8* undef, i8* null)
  unreachable

fpcall.next:                                      ; preds = %entry
  switch i32 %1, label %func.default [
    i32 0, label %func.nil
    i32 1, label %func.call1
    i32 2, label %func.call2
  ]

func.nil:                                         ; preds = %fpcall.next
  call void @runtime.nilPanic(i8* undef, i8* null)
  unreachable

func.call1:                                       ; preds = %fpcall.next
  call void @func1Uint8(i8 %2, i8* %0, i8* undef)
  br label %func.next

func.call2:                                       ; preds = %fpcall.next
  call void @func2Uint8(i8 %2, i8* %0, i8* undef)
  br label %func.next

func.next:                                        ; preds = %func.call2, %func.call1
  ret void

func.default:                                     ; preds = %fpcall.next
  unreachable
}

define void @sleepFuncValue(i8* %0, i32 %1, i8* nocapture readnone %context, i8* nocapture readnone %parentHandle) {
entry:
  switch i32 %1, label %func.default [
    i32 0, label %func.nil
    i32 1, label %func.call1
    i32 2, label %func.call2
  ]

func.nil:                                         ; preds = %entry
  call void @runtime.nilPanic(i8* undef, i8* null)
  unreachable

func.call1:                                       ; preds = %entry
  call void @"internal/task.start"(i32 ptrtoint (void (i32, i8*, i8*)* @"main$1" to i32), i8* null, i8* undef, i8* null)
  br label %func.next

func.call2:                                       ; preds = %entry
  call void @"internal/task.start"(i32 ptrtoint (void (i32, i8*, i8*)* @"main$2" to i32), i8* null, i8* undef, i8* null)
  br label %func.next

func.next:                                        ; preds = %func.call2, %func.call1
  ret void

func.default:                                     ; preds = %entry
  unreachable
}
