extern crate alloc;
extern crate core;
extern crate wee_alloc;

use alloc::vec::Vec;
use std::mem::MaybeUninit;
use std::slice;

// ##################
// # BUSINESS LOGIC #
// ##################

fn hello_wasm(name: &String) -> String {
    return ["hi ", &name].concat();
}

// #########################
// # WEBASSEMBLY INTERNALS #
// #########################

#[global_allocator]
static ALLOC: wee_alloc::WeeAlloc = wee_alloc::WeeAlloc::INIT;

unsafe fn ptr_to_string(ptr: u32, size: u32) -> String {
    let slice = slice::from_raw_parts_mut(ptr as *mut u8, size as usize);
    let utf8 = std::str::from_utf8_unchecked_mut(slice);
    return String::from(utf8);
}

// ##################
// # MAIN FUNCTIONS #
// ##################

#[no_mangle]
pub extern "C" fn allocate(size: u32) -> u32 {
    let vec: Vec<MaybeUninit<u32>> = vec![MaybeUninit::uninit(); size as usize];
    let ptr = Box::into_raw(vec.into_boxed_slice()) as *mut u32;
    return ptr as u32;
}

#[no_mangle]
pub unsafe extern "C" fn deallocate(ptr: u32, size: u32) {
    let _ = Vec::from_raw_parts(ptr as *mut u32, 0, size as usize);
}

#[no_mangle]
pub unsafe extern "C" fn boot(ptr: u32, size: u32) -> u64 {
    let inlet = ptr_to_string(ptr, size);
    let result = hello_wasm(&inlet);
    let leaked: &'static mut str = Box::leak(result.into_boxed_str());
    let (outlet, outlet_size) = (leaked.as_ptr() as u32, leaked.len() as u32);
    return ((outlet as u64) << 32) | (outlet_size as u64);
}

