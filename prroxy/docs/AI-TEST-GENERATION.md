# AI-Assisted Test Case Generation

## Overview

Use AI to automatically generate comprehensive test cases for API migration testing - just point AI at your code folders and let it discover everything.

## The Single Prompt Approach

**Just ask AI to read your folders and generate everything in one go:**

```
Please read all files in my src/api/ folder (or wherever your API code lives).

Also check if there's:
- A swagger.json or openapi.yaml file
- Any README or docs/ folder with API documentation
- Any example request/response files
- Any test data or sample IDs/names in the code

Then generate a comprehensive test-cases.json file that:
- Covers all API endpoints you found
- Uses realistic test data (if you see example names/IDs, use similar ones)
- Includes successful scenarios and error cases (404, 400, 500)
- Tests edge cases (invalid IDs, missing parameters, boundary conditions)
- Saves as test-data/test-cases.json

Format each test case as:
{
  "test_cases": [
    {
      "id": "unique_test_id",
      "name": "Descriptive test name",
      "endpoint": "/api/path",
      "method": "GET|POST|PUT|DELETE",
      "headers": {},
      "request_body": {},
      "expected_response": {}
    }
  ]
}

If you find example data in the code (user names, IDs, etc.), please use
similar realistic data in all test cases.

After saving, please validate the JSON file by running:
  ./scripts/validate-test-cases.sh

This ensures the file has valid syntax and all required fields.
```

**That's it!** AI will:
- Explore your folders
- Find all relevant files (code, swagger, docs, examples)
- Extract endpoint definitions
- Discover example data
- Generate comprehensive tests
- Save to the correct location

## What AI Will Discover

AI can automatically find and use:

- **Source Code** - API route handlers, controllers, services
- **Swagger/OpenAPI** - API specifications
- **Documentation** - README files, API docs
- **Example Data** - Sample IDs, names, test values in comments or test files
- **Route Definitions** - All endpoints and their parameters
- **Database Models** - Understanding data structures

## After AI Generates Tests

1. **Validate** - Check the generated file is valid:
   ```bash
   ./scripts/validate-test-cases.sh
   ```
   This checks for:
   - Valid JSON syntax
   - Required fields (id, name, endpoint, method)
   - No duplicate test IDs
   - Proper structure

2. **Review** - Check the test-data/test-cases.json file

3. **Refine if needed** - Ask AI to add more edge cases:
   ```
   Please add test cases for:
   - Very large IDs (999999)
   - Negative IDs (-1)
   - String IDs instead of numbers ("abc")
   ```

4. **Run Recording** - Capture real API responses:
   ```bash
   ./scripts/execute-and-record-v1.sh
   ```

File location:
```
prroxy/
└── test-data/
    └── test-cases.json  ← AI saves here automatically
```

## Iterative Improvement

After running comparison, if coverage is low:

```
I ran the tests and got 60% API coverage.

Looking at my code, I also see:
- Pagination parameters (?page=1&limit=10)
- Sorting options (?sort=name&order=asc)
- Filter parameters (?status=active)

Please read my current test-data/test-cases.json and add test cases
for these features. Keep using realistic test data.
```

## Tips for Best Results

1. **Let AI explore** - Don't manually gather files, just point to folders
2. **Mention example data** - If you have specific IDs/names to use, say so
3. **Be specific about edge cases** - What errors matter to you?
4. **Iterate** - Start simple, add more based on coverage results
5. **Technology agnostic** - Works with any language/framework

## Example for Different Stacks

### REST API (any language)
```
Please read my src/api/ and docs/ folders.
Generate comprehensive test-cases.json for API migration testing.
Save to test-data/test-cases.json.
```

### GraphQL API
```
Please read my src/graphql/ folder and schema files.
Generate GraphQL query tests covering all resolvers.
Save to test-data/test-cases.json.
```

### gRPC API
```
Please read my proto/ folder with .proto files.
Generate test scenarios for all RPC methods.
Save to test-data/test-cases.json.
```

## Alternative: Manual Paste

If you prefer to paste specific content instead of having AI read folders:

```
I need comprehensive test cases. Here's what I have:

[Paste one or more:]
- Source code files
- Swagger/OpenAPI spec
- Example API calls
- Sample data (IDs, names, etc.)

Generate test-cases.json following the format above.
```

## What Makes Tests Better

**Realistic Data**
- Use actual names: "Sarah Chen", "Marcus Johnson"
- Not generic: "user1", "user2"

**Edge Cases**
- Invalid IDs (999, -1, "abc")
- Missing parameters
- Boundary values

**Coverage**
- All endpoints
- All parameter combinations
- Success and error scenarios

## Complete Workflow

1. **Point AI at folders** - One prompt with folder paths
2. **Validate** - Run `./scripts/validate-test-cases.sh`
3. **Review generated tests** - Check test-data/test-cases.json
4. **Refine if needed** - Ask for additional test cases
5. **Record tests** - Run recording scripts
6. **Compare** - Run comparison between old and new APIs
7. **Iterate** - Add more tests if coverage insufficient

---

**TL;DR**: One prompt: "Please read my src/api/ and docs/ folders, then generate comprehensive test-cases.json saved to test-data/test-cases.json". AI does all the discovery and test generation automatically.
