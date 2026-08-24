#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <assert.h>
#include "../bin/libvicore.h"

int main() {
    printf("=== Starting C ABI Test for libvicore ===\n");

    // 1. Test version
    char* version = vi_engine_version();
    printf("libvicore version: %s\n", version);
    assert(version != NULL);
    assert(strstr(version, "bamboo-viet") != NULL);

    // 2. Test Engine Creation (Telex)
    // flags: 7 (EstdFlags)
    uintptr_t handle = vi_engine_new("Telex", 7);
    assert(handle != 0);
    printf("Created Telex engine handle: %lu\n", handle);

    // 3. Test typing sequence: "t", "i", "e", "e", "n", "g", "s"
    const char* input_keys = "tieengs";
    char commit_buf[256] = {0};
    char preedit_buf[256] = {0};
    int bs_count = 0;

    for (int i = 0; i < strlen(input_keys); i++) {
        memset(commit_buf, 0, sizeof(commit_buf));
        memset(preedit_buf, 0, sizeof(preedit_buf));
        bs_count = 0;

        int processed = vi_engine_process_key(
            handle,
            (unsigned int)input_keys[i],
            (unsigned int)input_keys[i],
            0,
            commit_buf, sizeof(commit_buf),
            preedit_buf, sizeof(preedit_buf),
            &bs_count
        );

        assert(processed == 1);
        printf("Key '%c' -> Preedit: [%s], Commit: [%s], Backspace: %d\n",
               input_keys[i], preedit_buf, commit_buf, bs_count);
    }

    // Verify preedit is "tiếng"
    assert(strcmp(preedit_buf, "tiếng") == 0);
    printf("✓ Telex composition 'tieengs' -> 'tiếng' succeeded!\n");

    // 4. Test Space (Word break -> Commit before hide)
    memset(commit_buf, 0, sizeof(commit_buf));
    memset(preedit_buf, 0, sizeof(preedit_buf));
    int processed = vi_engine_process_key(
        handle,
        (unsigned int)' ',
        (unsigned int)' ',
        0,
        commit_buf, sizeof(commit_buf),
        preedit_buf, sizeof(preedit_buf),
        &bs_count
    );
    assert(processed == 1);
    assert(strcmp(commit_buf, "tiếng ") == 0);
    assert(strlen(preedit_buf) == 0);
    printf("✓ Word-break space committed: [%s]\n", commit_buf);

    // 5. Test VNI Engine
    uintptr_t vni_handle = vi_engine_new("VNI", 7);
    assert(vni_handle != 0);
    printf("Created VNI engine handle: %lu\n", vni_handle);

    const char* vni_keys = "vie6t5";
    for (int i = 0; i < strlen(vni_keys); i++) {
        memset(commit_buf, 0, sizeof(commit_buf));
        memset(preedit_buf, 0, sizeof(preedit_buf));
        vi_engine_process_key(
            vni_handle,
            (unsigned int)vni_keys[i],
            (unsigned int)vni_keys[i],
            0,
            commit_buf, sizeof(commit_buf),
            preedit_buf, sizeof(preedit_buf),
            &bs_count
        );
    }
    assert(strcmp(preedit_buf, "việt") == 0);
    printf("✓ VNI composition 'vie6t5' -> 'việt' succeeded!\n");

    // 6. Test Freeing Engine Instances
    vi_engine_free(handle);
    vi_engine_free(vni_handle);
    printf("✓ Freed engine handles successfully.\n");

    printf("=== All C ABI Tests Passed Successfully! ===\n");
    return 0;
}
