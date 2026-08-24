#ifndef _XTEST_H_
#define _XTEST_H_

#include <X11/Xlib.h>

#ifdef __cplusplus
extern "C" {
#endif

extern Bool XTestQueryExtension(
    Display*		/* dpy */,
    int*		/* event_base_return */,
    int*		/* error_base_return */,
    int*		/* major_version_return */,
    int*		/* minor_version_return */
);

extern Bool XTestCompareCursorWithWindow(
    Display*		/* dpy */,
    Window		/* window */,
    Cursor		/* cursor */
);

extern Bool XTestCompareCurrentCursorWithWindow(
    Display*		/* dpy */,
    Window		/* window */
);

extern int XTestFakeKeyEvent(
    Display*		/* dpy */,
    unsigned int	/* keycode */,
    Bool		/* is_press */,
    unsigned long	/* delay */
);

extern int XTestFakeButtonEvent(
    Display*		/* dpy */,
    unsigned int	/* button */,
    Bool		/* is_press */,
    unsigned long	/* delay */
);

extern int XTestFakeMotionEvent(
    Display*		/* dpy */,
    int			/* screen_number */,
    int			/* x */,
    int			/* y */,
    unsigned long	/* delay */
);

extern int XTestFakeRelativeMotionEvent(
    Display*		/* dpy */,
    int			/* x */,
    int			/* y */,
    unsigned long	/* delay */
);

extern int XTestGrabControl(
    Display*		/* dpy */,
    Bool		/* impervious */
);

extern void XTestSetGContextOfGC(
    GC			/* gc */,
    GContext		/* gid */
);

extern void XTestSetVisualIDOfVisual(
    Visual*		/* visual */,
    VisualID		/* visualid */
);

extern Status XTestDiscard(
    Display*		/* dpy */
);

#ifdef __cplusplus
}
#endif

#endif /* _XTEST_H_ */
