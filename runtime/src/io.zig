const std = @import("std");

pub fn naviary_print(value: c_int) void {
    std.debug.print("{d}\n", .{value});
}
