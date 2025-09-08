#include "runtime.h"
#include <cstdio>

extern "C" void navi_print_int(long long value) {
  printf("%lld\n", value);

  fflush(stdout);
}