#ifndef _RECORD_H_
#define _RECORD_H_

#include <X11/Xlib.h>

#ifdef __cplusplus
extern "C" {
#endif

#define XRecordFromServer               0
#define XRecordFromClient               1
#define XRecordClientStarted            2
#define XRecordClientDied               3
#define XRecordStartOfData              4
#define XRecordEndOfData                5

#define XRecordAllClients               3

typedef unsigned long XRecordContext;
typedef unsigned long XRecordClientSpec;

typedef struct {
    unsigned char       first;
    unsigned char       last;
} XRecordRange8;

typedef struct {
    unsigned short      first;
    unsigned short      last;
} XRecordRange16;

typedef struct {
    XRecordRange8       core_requests;
    XRecordRange8       core_replies;
    XRecordRange16      ext_requests;
    XRecordRange16      ext_replies;
    XRecordRange8       delivered_events;
    XRecordRange8       device_events;
    Bool                errors;
    Bool                client_started;
    Bool                client_died;
} XRecordRange;

typedef struct {
    XRecordContext      context;
    XRecordClientSpec   client;
    unsigned long       data_len;
    unsigned char*      data;
    int                 category;
    Bool                client_swapped;
    Time                server_time;
    unsigned long       recorded_sequence;
} XRecordInterceptData;

typedef void (*XRecordInterceptProc)(
    XPointer            /* closure */,
    XRecordInterceptData* /* hook */
);

extern Status XRecordQueryVersion(
    Display*            /* dpy */,
    int*                /* major_version */,
    int*                /* minor_version */
);

extern XRecordRange* XRecordAllocRange(void);

extern XRecordContext XRecordCreateContext(
    Display*            /* dpy */,
    int                 /* datum_flags */,
    XRecordClientSpec*  /* clients */,
    int                 /* nclients */,
    XRecordRange**      /* ranges */,
    int                 /* nranges */
);

extern Status XRecordEnableContext(
    Display*            /* dpy */,
    XRecordContext      /* context */,
    XRecordInterceptProc /* proc */,
    XPointer            /* closure */
);

extern Status XRecordDisableContext(
    Display*            /* dpy */,
    XRecordContext      /* context */
);

extern Status XRecordFreeContext(
    Display*            /* dpy */,
    XRecordContext      /* context */
);

extern void XRecordFreeData(
    XRecordInterceptData* /* data */
);

extern void XRecordProcessReplies(
    Display*            /* dpy */
);

#ifdef __cplusplus
}
#endif

#endif /* _RECORD_H_ */
