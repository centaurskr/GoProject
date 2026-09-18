#include <stdio.h>

// Define a simple C structure
typedef struct {
    int field1;
    float field2;
} MyCStruct;

// Function to return a C structure
MyCStruct returnStruct() {
    MyCStruct myStruct = {42, 3.14};
    return myStruct;
}

