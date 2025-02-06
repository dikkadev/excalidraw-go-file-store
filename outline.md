**Conceptual Outline for Reimplementing the Excalidraw Store Backend**

This outlines the architecture and functionality of a service designed to store and retrieve binary data. It acts as an intermediary API, handling requests from clients and interacting with an underlying data storage system.

**I. Core Functionality: Binary Data API**

The primary purpose is to provide a simple API to:

*   **Store Binary Data:** Accept binary data uploads from clients and persist them.
*   **Retrieve Binary Data:** Serve previously stored binary data to clients upon request.

**II. Key Components and Operations (Technology Agnostic):**

1.  **Configuration & Environment:**
    *   **Project Identification:** The service needs a way to identify its operational context, perhaps through a project name or environment settings. This could influence operational parameters.
    *   **Environment Distinction (Production vs. Development):**  The service should differentiate between production and development environments. This distinction affects aspects like data storage locations and security policies.
    *   **Data Container Name:**  A name needs to be configured for the storage location where data will be kept. Different names might be used for production and development. Think of this as a namespace or a top-level folder in your storage system.
    *   **Authentication for Local Development:** For local testing, a mechanism for authentication with the data storage system is required, potentially using a credentials file. Production environments might rely on different authentication methods, like service accounts or environment-provided credentials.
    *   **Data Size Limit:**  A maximum size limit for uploaded data needs to be defined and enforced.

2.  **Interaction with Data Storage System:**
    *   **Initialization:**  The service must establish a connection to the chosen data storage system at startup. In development, this might involve loading authentication information from a file.
    *   **Data Container Access:** Obtain a reference or access point to the configured data container within the storage system. All data will be stored and retrieved from within this container.

3.  **Web API Server (Conceptual):**
    *   **Server Setup:**  Initialize an HTTP server to handle incoming requests.
    *   **Request Handling:** The server must be capable of routing requests to appropriate handlers based on URL paths and HTTP methods (GET, POST).
    *   **Cross-Origin Resource Sharing (CORS) Management:** Implement a mechanism to control which web origins are permitted to access the API.
        *   **Lenient CORS for Data Retrieval (GET):**  Apply a less restrictive CORS policy for data retrieval requests, potentially allowing broader access.
        *   **Strict CORS for Data Upload (POST):** Implement a more restrictive CORS policy for data upload requests. This policy should validate the origin of the request against a predefined list of allowed origins.
    *   **Serving Static Assets (Optional):**  Optionally, the server can serve static files like a favicon and a basic HTML landing page.

4.  **API Endpoints (Conceptual):**
    *   **Data Retrieval Endpoint (`GET /api/v2/{dataKey}`):**
        *   **Method:** Handles HTTP `GET` requests.
        *   **Path Parameter:** Expects a `{dataKey}` in the URL path, which is the identifier of the data to retrieve.
        *   **CORS:** Applies the lenient CORS policy.
        *   **Processing Logic:**
            1.  Extract the `dataKey` from the URL.
            2.  Attempt to retrieve metadata for the data object associated with the `dataKey` from the storage system. This step is to verify if the data exists.
            3.  If the data exists:
                *   Set the HTTP response status to 200 (Success).
                *   Set the `Content-Type` header to `application/octet-stream` to indicate binary data.
                *   Create a readable data stream from the storage system for the given `dataKey`.
                *   Pipe this data stream as the response body to the client.
            4.  If the data does not exist (or an error occurs during retrieval):
                *   Log the error.
                *   Respond with a 404 HTTP status (Not Found) and a JSON response indicating failure (e.g., `{"message": "Could not find the file."}`).
    *   **Data Upload Endpoint (`POST /api/v2/post/`):**
        *   **Method:** Handles HTTP `POST` requests.
        *   **Path:** `/api/v2/post/`
        *   **CORS:** Applies the strict CORS policy with origin validation.
        *   **Processing Logic:**
            1.  Initialize a counter to track the size of the incoming data.
            2.  Generate a unique identifier (data key) for the new data object.
            3.  Create a writable data stream to the storage system using the generated data key. Configure it for non-resumable uploads (if desired).
            4.  **Error Handling for Data Stream:** Set up error handling for the write stream. If errors occur during writing to storage, log the error and respond with a 500 HTTP status (Internal Server Error) and a JSON error message (e.g., `{"message": "Upload error."}`).
            5.  **Success Handling for Data Stream:** Set up success handling for the write stream. When data writing completes successfully, respond with a 200 HTTP status (Success) and a JSON response containing:
                *   The generated `dataKey`.
                *   The full URL for retrieving the stored data via the data retrieval endpoint (`GET /api/v2/{dataKey}`). Construct this URL dynamically, considering whether the service is running in a secure (HTTPS) or local (HTTP) context and using the server's hostname from the request.
            6.  **Incoming Request Data Processing:** As data chunks are received from the client in the HTTP request body:
                *   Write each data chunk to the writable data stream, sending it to the storage system.
                *   Increment the data size counter by the size of the received chunk.
                *   **Data Size Limit Enforcement:** Check if the accumulated data size exceeds the configured limit. If it does:
                    *   Create an error object indicating that the data is too large and specifying the maximum allowed size.
                    *   Terminate the writable data stream to halt further writing.
                    *   Log the size limit error.
                    *   Immediately respond to the client with a 413 HTTP status (Payload Too Large) and a JSON response containing the error object.
            7.  **Request Completion Handling:** When the entire HTTP request body has been received from the client, finalize the writable data stream. This signals the storage system that the upload is complete and triggers the success or error handling of the stream.
            8.  **General Error Handling:** Enclose the entire data upload process in a general error handling block. If any unexpected error occurs during the upload process, catch it, log it, and respond with a 500 HTTP status (Internal Server Error) and a generic JSON error message (e.g., `{"message": "Could not upload the data."}`).

5.  **Server Startup and Listen:**
    *   Configure the port on which the HTTP server will listen for incoming requests. This port should be configurable via environment variables or a default value.
    *   Start the HTTP server and begin listening for requests on the configured port.
    *   Log a message to indicate that the service has started and is running, including the address (e.g., `http://localhost:{port}`) where it can be accessed locally.
