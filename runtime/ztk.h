#include <stdint.h>
#include <stdbool.h>
#include <string.h>

typedef void (*z_block_t)();

typedef struct {
    int  type_id;
    char data[];
} z_object_t;

typedef union {
    uintptr_t   integer;
    z_object_t *reference;
} z_value_t;

typedef struct {
    char       *name;
    z_value_t   call_impl;
    z_value_t   type_data;
    void      (*copy_impl)(z_object_t *ref);
    void      (*visit_impl)(z_object_t *ref);
} z_type_t;

z_type_t *z_type(z_value_t value);

bool      z_is_truthy(z_value_t value);
bool      z_extract_block(z_value_t value, z_block_t *block);

int       z_prolog(int varc);
z_value_t z_get_local(int id);
void      z_set_local(int id, z_value_t value);

int       z_argc();
void      z_push(z_value_t value);
z_value_t z_result();

void      z_return();
void      z_call();
void      z_return_call();

void      z_error(char *message);
