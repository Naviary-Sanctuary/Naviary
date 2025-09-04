pub mod garbage_collector;
pub mod object;

pub use garbage_collector::GarbageCollector;
pub use object::ObjectHeader;

use std::ffi::c_void;

// 전역 GC 인스턴스 (thread_local로 더 안전하게)
thread_local! {
    static GLOBAL_GC: std::cell::RefCell<Option<GarbageCollector>> = std::cell::RefCell::new(None);
}

// ===== C FFI 함수들 (LLVM이 호출) =====

#[unsafe(no_mangle)]
pub extern "C" fn naviary_runtime_init() -> *mut c_void {
    naviary_gc_init()
}

#[unsafe(no_mangle)]
pub extern "C" fn naviary_gc_init() -> *mut c_void {
    GLOBAL_GC.with(|gc| {
        let mut gc_ref = gc.borrow_mut();
        *gc_ref = Some(GarbageCollector::new());
        gc_ref.as_mut().unwrap() as *mut GarbageCollector as *mut c_void
    })
}

// Int 배열 할당
#[unsafe(no_mangle)]
pub extern "C" fn naviary_allocate_int_array(_gc: *mut c_void, capacity: usize) -> *mut c_void {
    GLOBAL_GC.with(|gc| {
        let mut gc_ref = gc.borrow_mut();
        if gc_ref.is_none() {
            *gc_ref = Some(GarbageCollector::new());
        }

        let garbage_collector = gc_ref.as_mut().unwrap();
        let array = garbage_collector.allocate_int_array(capacity);
        array as *mut c_void
    })
}

#[unsafe(no_mangle)]
pub extern "C" fn naviary_array_get_int(array: *mut c_void, index: usize) -> object::NaviaryInt {
    unsafe {
        let array = array as *mut object::IntArrayObject;
        (*array).get(index)
    }
}

#[unsafe(no_mangle)]
pub extern "C" fn naviary_array_set_int(
    array: *mut c_void,
    index: usize,
    value: object::NaviaryInt,
) {
    unsafe {
        let array = array as *mut object::IntArrayObject;

        // 배열 길이 확장 (필요시)
        if index >= (*array).length {
            (*array).length = index + 1;
        }

        (*array).set(index, value);
    }
}

// Float 배열 함수들
#[unsafe(no_mangle)]
pub extern "C" fn naviary_allocate_float_array(_gc: *mut c_void, capacity: usize) -> *mut c_void {
    GLOBAL_GC.with(|gc| {
        let mut gc_ref = gc.borrow_mut();
        if gc_ref.is_none() {
            *gc_ref = Some(GarbageCollector::new());
        }

        let garbage_collector = gc_ref.as_mut().unwrap();
        let array = garbage_collector.allocate_float_array(capacity);
        array as *mut c_void
    })
}

#[unsafe(no_mangle)]
pub extern "C" fn naviary_array_get_float(
    array: *mut c_void,
    index: usize,
) -> object::NaviaryFloat {
    unsafe {
        let array = array as *mut object::FloatArrayObject;
        (*array).get(index)
    }
}

#[unsafe(no_mangle)]
pub extern "C" fn naviary_array_set_float(
    array: *mut c_void,
    index: usize,
    value: object::NaviaryFloat,
) {
    unsafe {
        let array = array as *mut object::FloatArrayObject;

        if index >= (*array).length {
            (*array).length = index + 1;
        }

        (*array).set(index, value);
    }
}

// Bool 배열 함수들
#[unsafe(no_mangle)]
pub extern "C" fn naviary_allocate_bool_array(_gc: *mut c_void, capacity: usize) -> *mut c_void {
    GLOBAL_GC.with(|gc| {
        let mut gc_ref = gc.borrow_mut();
        if gc_ref.is_none() {
            *gc_ref = Some(GarbageCollector::new());
        }

        let garbage_collector = gc_ref.as_mut().unwrap();
        let array = garbage_collector.allocate_bool_array(capacity);
        array as *mut c_void
    })
}

#[unsafe(no_mangle)]
pub extern "C" fn naviary_array_get_bool(array: *mut c_void, index: usize) -> bool {
    unsafe {
        let array = array as *mut object::BoolArrayObject;
        (*array).get(index)
    }
}

#[unsafe(no_mangle)]
pub extern "C" fn naviary_array_set_bool(array: *mut c_void, index: usize, value: bool) {
    unsafe {
        let array = array as *mut object::BoolArrayObject;

        if index >= (*array).length {
            (*array).length = index + 1;
        }

        (*array).set(index, value);
    }
}

// String 배열 함수들
#[unsafe(no_mangle)]
pub extern "C" fn naviary_allocate_string_array(_gc: *mut c_void, capacity: usize) -> *mut c_void {
    GLOBAL_GC.with(|gc| {
        let mut gc_ref = gc.borrow_mut();
        if gc_ref.is_none() {
            *gc_ref = Some(GarbageCollector::new());
        }

        let garbage_collector = gc_ref.as_mut().unwrap();
        let array = garbage_collector.allocate_string_array(capacity);
        array as *mut c_void
    })
}

#[unsafe(no_mangle)]
pub extern "C" fn naviary_array_get_string(array: *mut c_void, index: usize) -> *mut c_void {
    unsafe {
        let array = array as *mut object::StringArrayObject;
        (*array).get(index) as *mut c_void
    }
}

#[unsafe(no_mangle)]
pub extern "C" fn naviary_array_set_string(array: *mut c_void, index: usize, value: *mut c_void) {
    unsafe {
        let array = array as *mut object::StringArrayObject;
        let string_obj = value as *mut object::StringObject;

        if index >= (*array).length {
            (*array).length = index + 1;
        }

        (*array).set(index, string_obj);
    }
}

// String 할당 (나중에 string 리터럴용)
#[unsafe(no_mangle)]
pub extern "C" fn naviary_allocate_string(text: *const u8, length: usize) -> *mut c_void {
    GLOBAL_GC.with(|gc| {
        let mut gc_ref = gc.borrow_mut();
        if gc_ref.is_none() {
            *gc_ref = Some(GarbageCollector::new());
        }

        let garbage_collector = gc_ref.as_mut().unwrap();

        unsafe {
            let slice = std::slice::from_raw_parts(text, length);
            let text_str = std::str::from_utf8_unchecked(slice);
            garbage_collector.allocate_string(text_str) as *mut c_void
        }
    })
}

// GC 실행
#[unsafe(no_mangle)]
pub extern "C" fn naviary_gc_collect(_gc: *mut c_void) {
    GLOBAL_GC.with(|gc| {
        if let Some(garbage_collector) = gc.borrow_mut().as_mut() {
            garbage_collector.collect();
        }
    });
}

// 루트 추가/제거
#[unsafe(no_mangle)]
pub extern "C" fn naviary_gc_add_root(_gc: *mut c_void, ptr: *mut c_void) {
    GLOBAL_GC.with(|gc| {
        if let Some(garbage_collector) = gc.borrow_mut().as_mut() {
            garbage_collector.add_root(ptr as *mut u8);
        }
    });
}

#[unsafe(no_mangle)]
pub extern "C" fn naviary_gc_remove_root(_gc: *mut c_void, ptr: *mut c_void) {
    GLOBAL_GC.with(|gc| {
        if let Some(garbage_collector) = gc.borrow_mut().as_mut() {
            garbage_collector.remove_root(ptr as *mut u8);
        }
    });
}
