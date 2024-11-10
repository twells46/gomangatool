In order to use the bubbletea tui, I will need to rework
everything I've already done to make them into tea.Cmds
Maybe not everything? Just the ones that I intended to be public.
In the examples, they run smaller/helper functions within that are not
part of the async tea.Cmd thing.

The first thing I want to tui is the functionality that I've aleardy
implemented.

Model 1 - textinput:
1. User inputs an ID
Stores result in shared state?
How to gracefully pass off data between models?

Model 2 - single select list, [tab] to submit button
---- Init function ----
2. Pull metadata
6. Store new tags and convert tag names to Tag struct
3. Parse the title options
---- Tea.cmd ----
4. User chooses title

Model 3 - textinput:
5. User inputs abbreviated title
Note: Should include chosen full title for reference

??
7. Combine to a Manga struct
8. Store in DB
