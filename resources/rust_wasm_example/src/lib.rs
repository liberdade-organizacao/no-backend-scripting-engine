extern crate alloc;
extern crate core;
extern crate wee_alloc;

use alloc::vec::Vec;
use std::mem::MaybeUninit;
use std::slice;
use wasm_bindgen::prelude::*;

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

unsafe fn ptr_to_string(ptr: u32, len: u32) -> String {
    let slice = slice::from_raw_parts_mut(ptr as *mut u8, len as usize);
    let utf8 = std::str::from_utf8_unchecked_mut(slice);
    return String::from(utf8);
}

unsafe fn string_to_ptr(s: &String) -> (u32, u32) {
    return (s.as_ptr() as u32, s.len() as u32);
}

// ##################
// # MAIN FUNCTIONS #
// ##################

#[wasm_bindgen]
pub fn allocate(size: u32) -> *mut u32 {
    let vec: Vec<MaybeUninit<u32>> = vec![MaybeUninit::uninit(); size as usize];
    return Box::into_raw(vec.into_boxed_slice()) as *mut u32;
}

#[wasm_bindgen]
pub unsafe fn deallocate(ptr: u32, size: u32) {
    let _ = Vec::from_raw_parts(ptr as *mut u32, 0, size as usize);
}

#[wasm_bindgen]
pub unsafe fn boot(ptr: u32, size: u32) -> u32 {
    let inlet = ptr_to_string(ptr, size);
    let result = hello_wasm(&inlet);
    let (outlet, _) = string_to_ptr(&result);
    return outlet;
}

