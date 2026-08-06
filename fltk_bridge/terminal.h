#pragma once

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef struct Fl_Terminal Fl_Terminal;
typedef struct GTerminal GTerminal;

GTerminal *go_fltk_new_Terminal(int x, int y, int w, int h, const char *label);
void go_fltk_Terminal_append(Fl_Terminal *terminal, const char *data, int length);
char *go_fltk_Terminal_text(Fl_Terminal *terminal, int lines_below_cursor);
void go_fltk_Terminal_clear(Fl_Terminal *terminal);
void go_fltk_Terminal_clear_history(Fl_Terminal *terminal);
void go_fltk_Terminal_reset(Fl_Terminal *terminal);
void go_fltk_Terminal_set_ansi(Fl_Terminal *terminal, int enabled);
int go_fltk_Terminal_ansi(Fl_Terminal *terminal);
void go_fltk_Terminal_set_history_rows(Fl_Terminal *terminal, int rows);
int go_fltk_Terminal_history_rows(Fl_Terminal *terminal);
int go_fltk_Terminal_display_rows(Fl_Terminal *terminal);
int go_fltk_Terminal_display_columns(Fl_Terminal *terminal);
int go_fltk_Terminal_fit_display_columns(GTerminal *terminal);
void go_fltk_Terminal_set_horizontal_scrollbar(Fl_Terminal *terminal, int style);
void go_fltk_Terminal_set_text_font(Fl_Terminal *terminal, int font);
void go_fltk_Terminal_set_text_size(Fl_Terminal *terminal, int size);
void go_fltk_Terminal_set_text_color(Fl_Terminal *terminal, unsigned int color);
void go_fltk_Terminal_set_background_color(Fl_Terminal *terminal, unsigned int color);
void go_fltk_Terminal_set_selection_colors(Fl_Terminal *terminal, unsigned int foreground, unsigned int background);
void go_fltk_Terminal_set_margins(Fl_Terminal *terminal, int left, int top, int right, int bottom);
void go_fltk_Terminal_set_redraw_rate(Fl_Terminal *terminal, float seconds);

#ifdef __cplusplus
}
#endif
