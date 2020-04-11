target datalayout = "e-m:e-p:32:32-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

%runtime.typecodeID = type { %runtime.typecodeID*, i32 }

@"reflect/types.type:basic:uint8" = external constant %runtime.typecodeID
@"reflect/types.type:basic:int" = external constant %runtime.typecodeID
@"func NeverImplementedMethod()" = external constant i8
@"Unmatched$interface" = private constant [1 x i8*] [i8* @"func NeverImplementedMethod()"]
@"func Double() int" = external constant i8
@"Doubler$interface" = private constant [1 x i8*] [i8* @"func Double() int"]
@"reflect/types.type:named:Number" = private constant %runtime.typecodeID { %runtime.typecodeID* @"reflect/types.type:basic:int", i32 0 }

declare i1 @runtime.interfaceImplements(i32, i8**)

declare i1 @runtime.typeAssert(i32, %runtime.typecodeID*)

declare i32 @runtime.interfaceMethod(i32, i8**, i8*)

declare void @runtime.printuint8(i8)

declare void @runtime.printint32(i32)

declare void @runtime.printptr(i32)

declare void @runtime.printnl()

define void @printInterfaces() {
  call void @printInterface(i32 4, i8* inttoptr (i32 5 to i8*))
  call void @printInterface(i32 16, i8* inttoptr (i8 120 to i8*))
  call void @printInterface(i32 68, i8* inttoptr (i32 3 to i8*))
  ret void
}

define void @printInterface(i32 %typecode, i8* %value) {
  %typeassert.ok1 = call i1 @"Unmatched$typeassert"(i32 %typecode)
  br i1 %typeassert.ok1, label %typeswitch.Unmatched, label %typeswitch.notUnmatched

typeswitch.Unmatched:                             ; preds = %0
  %unmatched = ptrtoint i8* %value to i32
  call void @runtime.printptr(i32 %unmatched)
  call void @runtime.printnl()
  ret void

typeswitch.notUnmatched:                          ; preds = %0
  %typeassert.ok = call i1 @"Doubler$typeassert"(i32 %typecode)
  br i1 %typeassert.ok, label %typeswitch.Doubler, label %typeswitch.notDoubler

typeswitch.Doubler:                               ; preds = %typeswitch.notUnmatched
  %doubler.result = call i32 @"(Number).Double$invoke"(i8* %value, i8* null)
  call void @runtime.printint32(i32 %doubler.result)
  ret void

typeswitch.notDoubler:                            ; preds = %typeswitch.notUnmatched
  %typeassert.ok2 = icmp eq i32 16, %typecode
  br i1 %typeassert.ok2, label %typeswitch.byte, label %typeswitch.notByte

typeswitch.byte:                                  ; preds = %typeswitch.notDoubler
  %byte = ptrtoint i8* %value to i8
  call void @runtime.printuint8(i8 %byte)
  call void @runtime.printnl()
  ret void

typeswitch.notByte:                               ; preds = %typeswitch.notDoubler
  ret void
}

define i32 @"(Number).Double"(i32 %receiver, i8* %parentHandle) {
  %ret = mul i32 %receiver, 2
  ret i32 %ret
}

define i32 @"(Number).Double$invoke"(i8* %receiverPtr, i8* %parentHandle) {
  %receiver = ptrtoint i8* %receiverPtr to i32
  %ret = call i32 @"(Number).Double"(i32 %receiver, i8* null)
  ret i32 %ret
}

define internal i1 @"Doubler$typeassert"(i32 %actualType) unnamed_addr {
entry:
  switch i32 %actualType, label %else [
    i32 68, label %then
  ]

then:                                             ; preds = %entry
  ret i1 true

else:                                             ; preds = %entry
  ret i1 false
}

define internal i1 @"Unmatched$typeassert"(i32 %actualType) unnamed_addr {
entry:
  switch i32 %actualType, label %else [
  ]

then:                                             ; No predecessors!
  ret i1 true

else:                                             ; preds = %entry
  ret i1 false
}
