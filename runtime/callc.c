#include "ztk.h"

typedef struct {
    int       code_pos;
    int       base_pos;
    int       value_pos;
    z_block_t block;
} z_frame_t;

#define STACK_SIZE 4096

static z_block_t block;
static int       code_pos;
static int       base_pos;
static int       value_pos;
static int       stack_pos;
static int       frame_pos;
static z_frame_t frames[1024];
static z_value_t locals[STACK_SIZE];

static void clear_state() {
    block = NULL;
    code_pos = 0;
    base_pos = 0;
    value_pos = 0;
    stack_pos = 0;
    frame_pos = 0;
}

z_value_t z_run(z_block_t prog) {
    clear_state();
    block = prog;
    while (block != NULL) {
        block();
    }
    return z_result();
}

int z_prolog(int varc) {
    if (code_pos != 0) {
        return code_pos;
    }
    base_pos = value_pos;
    value_pos = stack_pos + varc;
    stack_pos = value_pos;
    return 0;
}

z_value_t z_get_local(int id) {
    return locals[id + base_pos];
}

void z_set_local(int id, z_value_t value) {
    locals[id + base_pos] = value;
}

int z_argc() {
    return stack_pos - value_pos;
}

void check_stack() {
    if (stack_pos + 1 >= STACK_SIZE) {
        z_error("stack overflow");
    }
}

void z_push(z_value_t value) {
    locals[stack_pos++] = value;
}

z_value_t z_result() {
    return locals[value_pos];
}

void z_return() {
    if (frame_pos == 0) {
        locals[0] = locals[value_pos];
        clear_state();
        return;
    }
    z_frame_t f = frames[--frame_pos];
    locals[f.value_pos] = locals[value_pos];
    code_pos = f.code_pos;
    base_pos = f.base_pos;
    value_pos = f.value_pos;
    block = f.block;
    stack_pos = value_pos;
}

static void do_call() {
    z_block_t b;
    if (z_extract_block(locals[value_pos], &b)) {
        block = b;
        value_pos++;
        return;
    }

    for(;;) {
        z_value_t impl = z_type(locals[value_pos])->call_impl;
        if (!z_is_truthy(impl)) {
            z_error("not callable");
        }
        if (z_extract_block(impl, &b)) {
            block = b;
            return;
        }
        check_stack();
        memmove(locals + base_pos + 1, locals + base_pos, z_argc() * sizeof(z_value_t));
        stack_pos++;
        locals[base_pos] = impl;
    }
}

void z_call() {
    frames[frame_pos++] = (z_frame_t) {
        .code_pos  = code_pos,
        .base_pos  = base_pos,
        .value_pos = value_pos,
        .block     = block
    };
    code_pos = 0;
    do_call();
}

void z_return_call() {
    int argc = z_argc();
    memmove(locals + base_pos, locals + value_pos, argc * sizeof(z_value_t));
    value_pos = base_pos;
    stack_pos = value_pos + argc;
    code_pos = 0;
    do_call();
}


