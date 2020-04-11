target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32-unknown-unknown-wasm"

%runtime.stackChainObject = type { %runtime.stackChainObject*, i32 }

@runtime.stackChainStart = global %runtime.stackChainObject* null
@someGlobal = global i8 3

declare void @runtime.trackPointer(i8* nocapture readonly)

declare noalias nonnull i8* @runtime.alloc(i32)

define i8* @getPointer() {
  ret i8* @someGlobal
}

define i8* @needsStackSlots() {
  %gc.stackobject = alloca { %runtime.stackChainObject*, i32, i8* }
  store { %runtime.stackChainObject*, i32, i8* } { %runtime.stackChainObject* null, i32 1, i8* null }, { %runtime.stackChainObject*, i32, i8* }* %gc.stackobject
  %1 = load %runtime.stackChainObject*, %runtime.stackChainObject** @runtime.stackChainStart
  %2 = getelementptr { %runtime.stackChainObject*, i32, i8* }, { %runtime.stackChainObject*, i32, i8* }* %gc.stackobject, i32 0, i32 0
  store %runtime.stackChainObject* %1, %runtime.stackChainObject** %2
  %3 = bitcast { %runtime.stackChainObject*, i32, i8* }* %gc.stackobject to %runtime.stackChainObject*
  store %runtime.stackChainObject* %3, %runtime.stackChainObject** @runtime.stackChainStart
  %ptr = call i8* @runtime.alloc(i32 4)
  %4 = getelementptr { %runtime.stackChainObject*, i32, i8* }, { %runtime.stackChainObject*, i32, i8* }* %gc.stackobject, i32 0, i32 2
  store i8* %ptr, i8** %4
  store %runtime.stackChainObject* %1, %runtime.stackChainObject** @runtime.stackChainStart
  ret i8* %ptr
}

define i8* @needsStackSlots2() {
  %gc.stackobject = alloca { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8* }
  store { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8* } { %runtime.stackChainObject* null, i32 4, i8* null, i8* null, i8* null, i8* null }, { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8* }* %gc.stackobject
  %1 = load %runtime.stackChainObject*, %runtime.stackChainObject** @runtime.stackChainStart
  %2 = getelementptr { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8* }, { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8* }* %gc.stackobject, i32 0, i32 0
  store %runtime.stackChainObject* %1, %runtime.stackChainObject** %2
  %3 = bitcast { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8* }* %gc.stackobject to %runtime.stackChainObject*
  store %runtime.stackChainObject* %3, %runtime.stackChainObject** @runtime.stackChainStart
  %ptr1 = call i8* @getPointer()
  %4 = getelementptr { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8* }, { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8* }* %gc.stackobject, i32 0, i32 4
  store i8* %ptr1, i8** %4
  %5 = getelementptr { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8* }, { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8* }* %gc.stackobject, i32 0, i32 3
  store i8* %ptr1, i8** %5
  %6 = getelementptr { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8* }, { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8* }* %gc.stackobject, i32 0, i32 2
  store i8* %ptr1, i8** %6
  %ptr2 = getelementptr i8, i8* @someGlobal, i32 0
  %unused = call i8* @runtime.alloc(i32 4)
  %7 = getelementptr { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8* }, { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8* }* %gc.stackobject, i32 0, i32 5
  store i8* %unused, i8** %7
  store %runtime.stackChainObject* %1, %runtime.stackChainObject** @runtime.stackChainStart
  ret i8* %ptr1
}

define i8* @noAllocatingFunction() {
  %ptr = call i8* @getPointer()
  ret i8* %ptr
}

define i8* @fibNext(i8* %x, i8* %y) {
  %gc.stackobject = alloca { %runtime.stackChainObject*, i32, i8* }
  store { %runtime.stackChainObject*, i32, i8* } { %runtime.stackChainObject* null, i32 1, i8* null }, { %runtime.stackChainObject*, i32, i8* }* %gc.stackobject
  %1 = load %runtime.stackChainObject*, %runtime.stackChainObject** @runtime.stackChainStart
  %2 = getelementptr { %runtime.stackChainObject*, i32, i8* }, { %runtime.stackChainObject*, i32, i8* }* %gc.stackobject, i32 0, i32 0
  store %runtime.stackChainObject* %1, %runtime.stackChainObject** %2
  %3 = bitcast { %runtime.stackChainObject*, i32, i8* }* %gc.stackobject to %runtime.stackChainObject*
  store %runtime.stackChainObject* %3, %runtime.stackChainObject** @runtime.stackChainStart
  %x.val = load i8, i8* %x
  %y.val = load i8, i8* %y
  %out.val = add i8 %x.val, %y.val
  %out.alloc = call i8* @runtime.alloc(i32 1)
  %4 = getelementptr { %runtime.stackChainObject*, i32, i8* }, { %runtime.stackChainObject*, i32, i8* }* %gc.stackobject, i32 0, i32 2
  store i8* %out.alloc, i8** %4
  store i8 %out.val, i8* %out.alloc
  store %runtime.stackChainObject* %1, %runtime.stackChainObject** @runtime.stackChainStart
  ret i8* %out.alloc
}

define i8* @allocLoop() {
entry:
  %gc.stackobject = alloca { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8*, i8* }
  store { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8*, i8* } { %runtime.stackChainObject* null, i32 5, i8* null, i8* null, i8* null, i8* null, i8* null }, { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8*, i8* }* %gc.stackobject
  %0 = load %runtime.stackChainObject*, %runtime.stackChainObject** @runtime.stackChainStart
  %1 = getelementptr { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8*, i8* }, { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8*, i8* }* %gc.stackobject, i32 0, i32 0
  store %runtime.stackChainObject* %0, %runtime.stackChainObject** %1
  %2 = bitcast { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8*, i8* }* %gc.stackobject to %runtime.stackChainObject*
  store %runtime.stackChainObject* %2, %runtime.stackChainObject** @runtime.stackChainStart
  %entry.x = call i8* @runtime.alloc(i32 1)
  %3 = getelementptr { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8*, i8* }, { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8*, i8* }* %gc.stackobject, i32 0, i32 2
  store i8* %entry.x, i8** %3
  %entry.y = call i8* @runtime.alloc(i32 1)
  %4 = getelementptr { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8*, i8* }, { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8*, i8* }* %gc.stackobject, i32 0, i32 3
  store i8* %entry.y, i8** %4
  store i8 1, i8* %entry.y
  br label %loop

loop:                                             ; preds = %loop, %entry
  %prev.y = phi i8* [ %entry.y, %entry ], [ %prev.x, %loop ]
  %prev.x = phi i8* [ %entry.x, %entry ], [ %next.x, %loop ]
  %5 = getelementptr { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8*, i8* }, { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8*, i8* }* %gc.stackobject, i32 0, i32 5
  store i8* %prev.y, i8** %5
  %6 = getelementptr { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8*, i8* }, { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8*, i8* }* %gc.stackobject, i32 0, i32 4
  store i8* %prev.x, i8** %6
  %next.x = call i8* @fibNext(i8* %prev.x, i8* %prev.y)
  %7 = getelementptr { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8*, i8* }, { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8*, i8* }* %gc.stackobject, i32 0, i32 6
  store i8* %next.x, i8** %7
  %next.x.val = load i8, i8* %next.x
  %loop.done = icmp ult i8 40, %next.x.val
  br i1 %loop.done, label %end, label %loop

end:                                              ; preds = %loop
  store %runtime.stackChainObject* %0, %runtime.stackChainObject** @runtime.stackChainStart
  ret i8* %next.x
}
