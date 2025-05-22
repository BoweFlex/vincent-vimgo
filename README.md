# Vincent VimGo

An attempt at writing a code editor in Golang, eventually to implement Vim motions. 

## TODO

I think these two should be done in roughly this order:

- Paging to create a "window" of text
    - I'm picturing a horizontal offset and a vertical offset, and only lines above the vertical offset are shown and only columns above the horizontal offset are shown
    - Current logic to only show up to w and up to h are fine, but starting bounds are needed
- Concurrent screen drawing
    - "Buffer" and status line should be drawn simultaneously, and cursor should either also be concurrent or should be a part of drawing the buffer

These are all needed but probably not in this order:

- More Commands, currently only :q is supported.
- Reading in/writing to files
    - This may be an issue for down the line but how do you performantly handle very large files?
- Rope data structure, this is partially implemented in `github.com/BoweFlex/data-structures`
- Entering text at cursor's current location
    - Currently any text added is added at the end of the "buffer" regardless of position
- More support for vim motions
    - i.e. nums before actions in normal mode, jumping to end of line or beginning of line
- Line Numbers
- Configuration (temporary and permanent via conf file, can we integrate XDG?)
- Deterministic Simulation Testing
    - Ideally this would have a "fuzzer" and be able to simulate a variety of performance scenarios and input speeds
    - This would be nice to have, and a great skill to practice, but not sure it's more important than more concrete progress, at this point having a variety of skills and a somewhat finished product is likely better than an extremely stable half finished project.

Things that would be nice to have but are less of a priority than solid and basic editing of a single file:

- Directory Navigation
    - Not sure if this should be something like netrw, fuzzy finding like helix, or both
- Multiple buffers
    - This would also unlock the ability for split screening.
- Treesitter/tokenizing
    - Also unlocks the ability for additional vim motions
- LSP support
- Colorschemes
