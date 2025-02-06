**Integration Test Definitions for Excalidraw Store Backend**

We will categorize the tests by the API endpoint being tested and then by success and error scenarios.

**I. Tests for `POST /api/v2/post/` (Data Upload)**

**A. Successful Uploads:**

1.  **Test Case: Basic Successful Upload - Small Payload**
    *   **Description:** Verifies that a small binary payload can be successfully uploaded and the API returns a valid response.
    *   **Test Steps:**
        1.  Prepare a small binary payload (e.g., a few kilobytes of random data).
        2.  Send a `POST` request to `/api/v2/post/` with the payload in the request body. Ensure the `Origin` header is set to a valid allowed origin (e.g., `https://excalidraw.com`).
        3.  Wait for the API to process the request and send a response.
    *   **Expected Outcome:**
        *   HTTP response status code: `200 OK`.
        *   Response body is a JSON object with the following structure:
            ```json
            {
              "id": "[valid data key]",
              "data": "[valid retrieval URL for the data key]"
            }
            ```
        *   The `id` in the response should be a non-empty string (a valid data key).
        *   The `data` in the response should be a valid URL pointing to the `GET /api/v2/{dataKey}` endpoint, using the generated `id`.
        *   The uploaded binary payload should be successfully stored in the data storage system, accessible using the returned `dataKey`.
    *   **Failure Condition:**
        *   HTTP status code is not `200`.
        *   Response body is not in the expected JSON format.
        *   Response body does not contain `id` or `data` fields, or they are invalid.
        *   Data is not found in the storage system using the returned `dataKey` (verify this by attempting a `GET` request in a subsequent test or as part of this test).

2.  **Test Case: Successful Upload - Maximum Allowed Payload Size**
    *   **Description:** Checks if uploading a payload exactly at the maximum allowed size limit is successful.
    *   **Test Steps:**
        1.  Prepare a binary payload exactly equal to the configured `FILE_SIZE_LIMIT`.
        2.  Send a `POST` request to `/api/v2/post/` with this payload.  Set a valid `Origin` header.
        3.  Wait for the API response.
    *   **Expected Outcome:** Same as "Basic Successful Upload - Small Payload" - successful 200 response with valid data key and retrieval URL, and the data stored correctly.
    *   **Failure Condition:** Same as "Basic Successful Upload - Small Payload" - any deviation from the expected success response or data storage.

**B. Error Handling during Uploads:**

3.  **Test Case: Upload Payload Too Large - Exceeds Size Limit**
    *   **Description:** Verifies that the API correctly rejects uploads exceeding the maximum allowed size.
    *   **Test Steps:**
        1.  Prepare a binary payload that is slightly larger than `FILE_SIZE_LIMIT`.
        2.  Send a `POST` request to `/api/v2/post/` with this oversized payload. Set a valid `Origin` header.
        3.  Wait for the API response.
    *   **Expected Outcome:**
        *   HTTP response status code: `413 Payload Too Large`.
        *   Response body is a JSON object with an error message, potentially including details about the size limit:
            ```json
            {
              "message": "Data is too large.",
              "max_limit": [FILE_SIZE_LIMIT value]
            }
            ```
        *   The oversized payload should **not** be stored in the data storage system.
    *   **Failure Condition:**
        *   HTTP status code is not `413`.
        *   HTTP status code is `200 OK` (indicating a successful upload despite being oversized).
        *   Response body does not contain the expected error message or `max_limit` information.
        *   Data is unexpectedly stored in the storage system.

4.  **Test Case: Invalid Origin - CORS Rejection**
    *   **Description:** Checks if the API correctly rejects `POST` requests from disallowed origins due to CORS policy.
    *   **Test Steps:**
        1.  Prepare a small binary payload.
        2.  Send a `POST` request to `/api/v2/post/` with the payload. Set the `Origin` header to an origin that is **not** in the allowed origins list (e.g., `http://untrusted-origin.com`).
        3.  Wait for the API response.
    *   **Expected Outcome:**
        *   HTTP response status code: Should be a CORS-related error.  The *exact* status code might depend on how CORS is implemented in your Go server (it might be something like `400 Bad Request`, or the request might be blocked by the browser/client before even reaching the server in some scenarios - in integration test context, we're likely testing the server's response when the request *does* reach it).  Crucially, it should **not** be `200 OK`.
        *   The server should **not** process the upload request.
        *   The payload should **not** be stored in the data storage system.
    *   **Failure Condition:**
        *   HTTP status code is `200 OK`, `201 Created`, or any other successful status.
        *   The server accepts the request and attempts to store the data from a disallowed origin.
        *   No CORS-related error is indicated in the response (or lack thereof if the request is blocked earlier).

**II. Tests for `GET /api/v2/{dataKey}` (Data Retrieval)**

**A. Successful Retrievals:**

5.  **Test Case: Successful Data Retrieval - Valid Data Key**
    *   **Description:** Verifies that data can be successfully retrieved using a valid data key.
    *   **Test Steps:**
        1.  **Pre-requisite:** First, perform a successful `POST /api/v2/post/` request (using Test Case 1 or 2) and obtain the `dataKey` from the response.
        2.  Send a `GET` request to `/api/v2/{dataKey}`, replacing `{dataKey}` with the key obtained in the previous step. Set a valid or any `Origin` header (as GET is likely more lenient with CORS).
        3.  Wait for the API response.
    *   **Expected Outcome:**
        *   HTTP response status code: `200 OK`.
        *   Response body is the **exact** binary payload that was originally uploaded in the `POST` request.
        *   `Content-Type` header in the response is `application/octet-stream`.
    *   **Failure Condition:**
        *   HTTP status code is not `200`.
        *   Response body is not the same binary data that was originally uploaded.
        *   `Content-Type` header is not `application/octet-stream`.

**B. Error Handling during Retrievals:**

6.  **Test Case: Data Not Found - Invalid Data Key**
    *   **Description:** Checks that the API returns a 404 error when attempting to retrieve data with a non-existent or invalid data key.
    *   **Test Steps:**
        1.  Choose a data key that is highly unlikely to exist in the storage system (e.g., a randomly generated UUID or a known non-existent key).
        2.  Send a `GET` request to `/api/v2/{invalidDataKey}`, replacing `{invalidDataKey}` with the chosen key.
        3.  Wait for the API response.
    *   **Expected Outcome:**
        *   HTTP response status code: `404 Not Found`.
        *   Response body is a JSON object with an error message:
            ```json
            {
              "message": "Could not find the file."
            }
            ```
    *   **Failure Condition:**
        *   HTTP status code is not `404`.
        *   HTTP status code is `200 OK` (incorrectly indicating success).
        *   Response body does not contain the expected error message or is not in JSON format.

**III. CORS for GET Requests (Verification)**

7.  **Test Case: GET Request from Allowed Origin (CORS Success)**
    *   **Description:** Confirms that `GET` requests from allowed origins are successful (specifically focusing on CORS aspect).
    *   **Test Steps:**
        1.  Pre-requisite: Upload data and get a valid `dataKey` (like in Test Case 5).
        2.  Send a `GET` request to `/api/v2/{dataKey}`. Set the `Origin` header to a valid allowed origin (e.g., `https://excalidraw.com`).
        3.  Wait for the API response.
    *   **Expected Outcome:** Same successful outcome as Test Case 5 (200 OK, correct data, etc.). This primarily confirms CORS doesn't block valid GET requests.

8.  **Test Case: GET Request from Disallowed Origin (CORS - Should Still Succeed but Verify Headers)**
    *   **Description:**  Verify that `GET` requests from *disallowed* origins still succeed in retrieving data (as `corsGet` is intended to be lenient), but check for appropriate CORS headers (or lack thereof, depending on desired lenient policy).  *Note: Based on the code, `corsGet = cors()` which might be configured to allow all origins. Adjust this test and expected outcome based on the actual CORS policy you intend to implement for GET requests.*  If you intend to restrict GET CORS, adjust the `corsGet` setup and this test accordingly.
    *   **Test Steps:**
        1.  Pre-requisite: Upload data and get a valid `dataKey`.
        2.  Send a `GET` request to `/api/v2/{dataKey}`. Set the `Origin` header to a **disallowed** origin (e.g., `http://untrusted-origin.com`).
        3.  Wait for the API response.
    *   **Expected Outcome (if lenient GET CORS - like `cors()` defaults):**
        *   HTTP response status code: `200 OK`.
        *   Response body is the correct data.
        *   **Crucially:** Check the CORS headers in the response. For a truly open GET policy, you might expect headers like `Access-Control-Allow-Origin: *` (if that's how your CORS middleware is configured). If you have a more specific lenient policy, check for the headers you expect based on that policy.  If you intend *no* CORS for GET, then no `Access-Control-Allow-Origin` header should be present (or it might be absent by default if no CORS middleware is explicitly applied - check your implementation).

**General Notes for Implementation:**

*   **Test Environment:** Set up a dedicated test environment for your integration tests, ideally mimicking your production environment setup as closely as possible (e.g., using a test data storage container).
*   **Test Data Generation:**  Create helper functions to generate binary payloads of different sizes and content for your tests.
*   **HTTP Client:** Use a suitable HTTP client library in your testing framework to send requests to your API.
*   **Assertions:** Use assertion libraries to verify the expected outcomes (status codes, response bodies, headers).
*   **Data Key Management:**  Be mindful of data keys generated during tests. You might need to implement cleanup mechanisms in your tests to remove test data from the storage system after tests are run, to prevent accumulation of test data.
