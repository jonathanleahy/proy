Feature: HTTP Testing Proxy
  As a tester
  I want to record and replay HTTP interactions
  So that I can test without depending on external services

  Background:
    Given the proxy is running on port 8080
    And the recordings directory is empty

  Scenario: Record a new HTTP interaction
    Given the proxy is in "record" mode
    When I send a GET request to "/proxy?target=api.example.com/users"
    Then the request should be forwarded to "api.example.com"
    And the response should be returned to the client
    And the interaction should be saved in the recordings directory

  Scenario: Replay a recorded HTTP interaction
    Given the proxy is in "record" mode
    And I send a GET request to "/proxy?target=api.example.com/users" with response "{"users":[]}"
    And the proxy is switched to "playback" mode
    When I send a GET request to "/proxy?target=api.example.com/users"
    Then the recorded response "{"users":[]}" should be returned
    And no external request should be made

  Scenario: Fail fast when no recording exists
    Given the proxy is in "playback" mode
    When I send a GET request to "/proxy?target=api.example.com/unknown"
    Then the proxy should return a 404 error
    And the error message should contain "No recording found"

  Scenario: Full request matching
    Given the proxy is in "record" mode
    And I send a POST request to "/proxy?target=api.example.com/users" with:
      | header        | value            |
      | Content-Type  | application/json |
      | X-Tenant      | org-123          |
      | body          | {"name":"Alice"} |
    And the proxy is switched to "playback" mode
    When I send a POST request to "/proxy?target=api.example.com/users" with:
      | header        | value            |
      | Content-Type  | application/json |
      | X-Tenant      | org-123          |
      | body          | {"name":"Alice"} |
    Then the recorded response should be returned
    When I send a POST request to "/proxy?target=api.example.com/users" with:
      | header        | value            |
      | Content-Type  | application/json |
      | X-Tenant      | org-456          |
      | body          | {"name":"Alice"} |
    Then the proxy should return a 404 error

  Scenario: Switch between modes via API
    Given the proxy is in "record" mode
    When I send a POST request to "/admin/mode" with body "{"mode":"playback"}"
    Then the response should contain "{"mode":"playback"}"
    And the proxy should be in "playback" mode
    When I send a POST request to "/admin/mode" with body "{"mode":"record"}"
    Then the response should contain "{"mode":"record"}"
    And the proxy should be in "record" mode

  Scenario: List recorded interactions
    Given the proxy is in "record" mode
    And I have recorded 3 different requests
    When I send a GET request to "/admin/recordings"
    Then the response should list 3 recordings
    And each recording should have an ID, timestamp, and request details

  Scenario: Clear all recordings
    Given the proxy is in "record" mode
    And I have recorded 5 different requests
    When I send a DELETE request to "/admin/recordings"
    Then the response should confirm deletion
    And the recordings directory should be empty
    When I send a GET request to "/admin/recordings"
    Then the response should list 0 recordings

  Scenario: View statistics
    Given the proxy is in "playback" mode
    And I have recorded 10 different requests
    And I replay 3 existing requests
    And I request 2 non-existing endpoints
    When I send a GET request to "/admin/status"
    Then the response should show:
      | field            | value    |
      | mode             | playback |
      | total_recordings | 10       |
      | cache_hits       | 3        |
      | cache_misses     | 2        |

  Scenario: Handle query parameters in target URL
    Given the proxy is in "record" mode
    When I send a GET request to "/proxy?target=api.example.com/users?page=1&limit=10"
    Then the request should be forwarded to "api.example.com/users?page=1&limit=10"
    And the full URL with query parameters should be recorded

  Scenario: Organize recordings by service
    Given the proxy is in "record" mode
    When I send requests to different services:
      | target                    | method |
      | api.users.com/list        | GET    |
      | api.accounts.com/balance  | GET    |
      | api.users.com/create      | POST   |
    Then the recordings should be organized in directories:
      | service           | count |
      | api_users_com     | 2     |
      | api_accounts_com  | 1     |