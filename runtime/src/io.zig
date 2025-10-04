const std = @import("std");

pub fn naviary_print(value: i64) void {
    std.debug.print("{d}\n", .{value});
}
