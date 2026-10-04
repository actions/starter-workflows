#include <stdio.h>

int add(int a, int b);

int main() {
    if (add(2, 3) != 5) {
        printf("Test failed: 2 + 3 should equal 5\n");
        return 1;
    }

    if (add(10, 20) != 30) {
        printf("Test failed: 10 + 20 should equal 30\n");
        return 1;
    }

    printf("All tests passed!\n");
    return 0;
}
