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


Next task: the main gui.
Should be a list that displays the manga in your library.
When you select a series, you can then look through the chapters in that series.
That means implementing list.Item for both Chapter and Manga.
Also want keybind to refresh series.
Have 'a' keybind to add series.

The overall view should show the title, tags, and a snippet of the description.
Also maybe update time?
Should the tags and demographic and other filter stuff be part of the builtin filter for lists?
Probably not.

Have a seperate thing with a couple of lists to filter series.

The individual series view should show the title, tags, full description, update time.
Has the list of chapters below.

Still not sure about how I want the code organized.
