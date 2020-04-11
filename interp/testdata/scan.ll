target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

%runtime.typecodeID = type { %runtime.typecodeID*, i64 }

declare i1 @runtime.typeAssert(i64, %runtime.typecodeID*, i8*, i8*)
declare i1 @runtime.interfaceImplements(i64, i8**)

define i64 @returnsConst() {
  ret i64 0
}

define i64 @returnsArg(i64 %arg) {
  ret i64 %arg
}

declare i64 @externalCall()

define i64 @externalCallOnly() {
  %result = call i64 @externalCall()
  ret i64 0
}

define i64 @externalCallAndReturn() {
  %result = call i64 @externalCall()
  ret i64 %result
}

define i64 @externalCallBranch() {
  %result = call i64 @externalCall()
  %zero = icmp eq i64 %result, 0
  br i1 %zero, label %if.then, label %if.done

if.then:
  ret i64 2

if.done:
  ret i64 4
}

@cleanGlobalInt = global i64 5
define i64 @readCleanGlobal() {
  %global = load i64, i64* @cleanGlobalInt
  ret i64 %global
}

@dirtyGlobalInt = global i64 5
define i64 @readDirtyGlobal() {
  %global = load i64, i64* @dirtyGlobalInt
  ret i64 %global
}

declare i64* @getDirtyPointer()

define void @storeToPointer() {
  %ptr = call i64* @getDirtyPointer()
  store i64 3, i64* %ptr
  ret void
}

@functionPointer = global i64()* null
define i64 @callFunctionPointer() {
  %fp = load i64()*, i64()** @functionPointer
  %result = call i64 %fp()
  ret i64 %result
}

define i1 @callTypeAssert() {
  ; Note: parameters are not realistic.
  %ok = call i1 @runtime.typeAssert(i64 0, %runtime.typecodeID* null, i8* undef, i8* null)
  ret i1 %ok
}

define i1 @callInterfaceImplements() {
  ; Note: parameters are not realistic.
  %ok = call i1 @runtime.interfaceImplements(i64 0, i8** null)
  ret i1 %ok
}
