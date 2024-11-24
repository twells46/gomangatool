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

I want to move to having the main gui.go file call a single view and update function per "thing".
i.e. One update function for adder, one for all, one for series.
Then those can call more subviews as necessary.
This will allow me to have shared keybinds and UI elements between them more easily.

Change the styles based on new chapters

Add review mechanism

Heres the current problem:
When I click into a series, it needs to load the chapters so the user can view them.
To do this, I have a flag which, when false, calls the function to do the loading.
This function seems to be basically identical to the getTitles function from adder, but for some reason it behaves incorrectly.
It does not update the view until the user presses a key.
I don't have this problem anywhere else, so I'm really not sure why it is such a hassle here.
