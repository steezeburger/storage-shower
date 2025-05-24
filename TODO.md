Feature Requests 5/23:
- [ ] the user should see a legend to describe what the different colors used represent
- [ ] chart should be zoomable so i can see the whole thing for large starbursts
- [ ] should be a `>` between breadcrumb chips

Regressions:
- [x] some scans showing 0 bytes again, like /Users/jessesnyder/code/retro-gfx/, but it works fine for /Users/jessesnyder/code/projects/archived. /Users/jessesnyder/code/projects/storage-shower/ is broken though. related to dashes?
- [x] server no longer runs on port 8080
- [x] localhost/:1 Refused to execute script from 'http://localhost:8080/frontend/app.js' because its MIME type ('text/plain') is not executable, and strict MIME type checking is enabled.

Bugs:
- [ ] /scans 404s
- [x] the debug logs are always shown regardless of flag existence
- [x] clicking into directory view from main view treemap errors
  ```
  d3.v7.min.js:2 Error: <rect> attribute height: Expected length, "NaN".
  (anonymous)	@	d3.v7.min.js:2
  each	@	d3.v7.min.js:2
  attr	@	d3.v7.min.js:2
  renderTreemap	@	app.js:463
  renderVisualization	@	app.js:379
  (anonymous)	@	app.js:450
  (anonymous)	@	d3.v7.min.js:2
  ```
- [x] the stalled button is sometimes incorrectly shown

Feature Requests 5/22:
- [x] i want to click the filepath on web to copy the full path to my clipboard
- [x] i want to see previous runs in the web underneath my current run
- [x] the project should have a gitignore with relevant files and dirs included for this project
- [x] the project should setup claude code github actions following instructions fromhttps://docs.anthropic.com/en/docs/claude-code/github-actions
- [x] i want to see the logo from images/logo.jpeg in the README.md but please convert the file to a png first
- [x] i do not want to ignore hidden files by default
- [x] i want the browse button to open up a file explorer so i can select a directory
- [x] the project should have a code formatter. use defaults of the most popular formatter for the language.
- [x] the project should have a linter
- [x] move the build script logic to the justfile
- [x] the projet should have a github action to check the formatting and run the linter
- [x] the project should have backend tests
- [x] the project should have web tests
- [x] the project should have a github action to run the tests
