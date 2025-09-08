.text
.align 2
.globl _main

_main:
    stp     x29, x30, [sp, #-16]!   
    mov     x29, sp                  
    
    mov     x0, #42                  
    bl      _navi_print_int          
    
    mov     x0, #100
    bl      _navi_print_int
    
    mov     x0, #-7
    bl      _navi_print_int
    
    mov     x0, #0                   
    
    ldp     x29, x30, [sp], #16      
    ret                              