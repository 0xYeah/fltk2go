#include "terminal.h"

#include <cstdlib>
#include <cstring>

#include <FL/Fl.H>
#include <FL/Fl_Terminal.H>

#include "event_handler.h"

class GTerminal : public EventHandler<Fl_Terminal> {
public:
  GTerminal(int x, int y, int w, int h, const char *label)
      : EventHandler<Fl_Terminal>(x, y, w, h, label) {}
};

GTerminal *go_fltk_new_Terminal(int x, int y, int w, int h, const char *label) {
  return new GTerminal(x, y, w, h, label);
}

void go_fltk_Terminal_append(Fl_Terminal *terminal, const char *data, int length) {
  if (terminal && data && length > 0) terminal->append(data, length);
}

char *go_fltk_Terminal_text(Fl_Terminal *terminal, int lines_below_cursor) {
  const char *text = terminal ? terminal->text(lines_below_cursor != 0) : "";
  if (!text) text = "";
  const size_t length = std::strlen(text);
  char *copy = static_cast<char *>(std::malloc(length + 1));
  if (!copy) return nullptr;
  std::memcpy(copy, text, length + 1);
  return copy;
}

void go_fltk_Terminal_clear(Fl_Terminal *terminal) {
  if (terminal) terminal->clear_screen_home(false);
}

void go_fltk_Terminal_clear_history(Fl_Terminal *terminal) {
  if (terminal) terminal->clear_history();
}

void go_fltk_Terminal_reset(Fl_Terminal *terminal) {
  if (terminal) terminal->reset_terminal();
}

void go_fltk_Terminal_set_ansi(Fl_Terminal *terminal, int enabled) {
  if (terminal) terminal->ansi(enabled != 0);
}

int go_fltk_Terminal_ansi(Fl_Terminal *terminal) {
  return terminal && terminal->ansi();
}

void go_fltk_Terminal_set_history_rows(Fl_Terminal *terminal, int rows) {
  if (terminal) terminal->history_rows(rows);
}

int go_fltk_Terminal_history_rows(Fl_Terminal *terminal) {
  return terminal ? terminal->history_rows() : 0;
}

int go_fltk_Terminal_display_rows(Fl_Terminal *terminal) {
  return terminal ? terminal->display_rows() : 0;
}

int go_fltk_Terminal_display_columns(Fl_Terminal *terminal) {
  return terminal ? terminal->display_columns() : 0;
}

void go_fltk_Terminal_set_text_font(Fl_Terminal *terminal, int font) {
  if (terminal) terminal->textfont(static_cast<Fl_Font>(font));
}

void go_fltk_Terminal_set_text_size(Fl_Terminal *terminal, int size) {
  if (terminal) terminal->textsize(size);
}

void go_fltk_Terminal_set_text_color(Fl_Terminal *terminal, unsigned int color) {
  if (!terminal) return;
  terminal->textcolor(static_cast<Fl_Color>(color));
  terminal->textfgcolor(static_cast<Fl_Color>(color));
}

void go_fltk_Terminal_set_background_color(Fl_Terminal *terminal, unsigned int color) {
  if (!terminal) return;
  terminal->color(static_cast<Fl_Color>(color));
  terminal->textbgcolor(static_cast<Fl_Color>(color));
  terminal->textbgcolor_default(static_cast<Fl_Color>(color));
}

void go_fltk_Terminal_set_selection_colors(Fl_Terminal *terminal, unsigned int foreground, unsigned int background) {
  if (!terminal) return;
  terminal->selectionfgcolor(static_cast<Fl_Color>(foreground));
  terminal->selectionbgcolor(static_cast<Fl_Color>(background));
}

void go_fltk_Terminal_set_margins(Fl_Terminal *terminal, int left, int top, int right, int bottom) {
  if (!terminal) return;
  terminal->margin_left(left);
  terminal->margin_top(top);
  terminal->margin_right(right);
  terminal->margin_bottom(bottom);
}

void go_fltk_Terminal_set_redraw_rate(Fl_Terminal *terminal, float seconds) {
  if (!terminal) return;
  terminal->redraw_style(Fl_Terminal::RATE_LIMITED);
  terminal->redraw_rate(seconds);
}
