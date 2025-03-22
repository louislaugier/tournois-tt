/*
Package helloasso provides functionality for searching and extracting activities from HelloAsso.

Testing Strategy:
----------------
The HelloAsso service interacts with the Playwright library for browser automation,
which makes it challenging to create traditional unit tests that mock these interactions.
We have adopted a logic-based testing approach:

1. Logic-based testing:
  - Instead of mocking the Playwright interfaces, we test the core logic
  - Helper functions simulate the behavior of the real implementations
  - Tests verify that given specific inputs (empty state, show all button, etc.),
    the expected outputs are produced
  - This approach focuses on business logic rather than implementation details
  - See `TestExtractActivitiesLogic` and `TestSearchActivitiesLogic` for examples

2. Why direct mocking was abandoned:
  - The original approach attempted to mock the Playwright interfaces
  - This proved challenging due to the large number of methods required by the interfaces
  - Type conversion issues occurred when trying to use the mock objects with the real code
  - It was decided to focus on behavior testing rather than implementation testing

For a complete test coverage, integration tests with a real browser instance would be valuable,
but those are not included in the unit test suite.
*/
package helloasso
