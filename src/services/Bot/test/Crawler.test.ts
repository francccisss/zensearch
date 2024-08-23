

test("Should say hello",()=>{
    const hello = "Hello";
    expect((function(hello){return hello})(hello)).toBe("Hello");
})