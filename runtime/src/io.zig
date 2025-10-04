const std = @import("std");

pub fn naviary_print_int(value: i64) void {
    std.debug.print("{d}\n", .{value});
}

pub fn naviary_print_string(string_pointer: [*:0]const u8) void {
    std.debug.print("{s}\n", .{string_pointer});
}
