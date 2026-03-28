III. Quality Acceptance Criteria
Important reminder : Before submitting the task, please be sure to self-test item by item against the following six dimensions.
3.1 and 3.2 are red line indicators : Once violated, it will be directly judged as "unqualified", and in principle, it will not enter the repair process and will be directly abandoned.
3.3 to 3.6 are quality indicators : affecting the score and the number of repairs, and repeated failure to meet the standard will affect the dispatch of follow-up tasks.
When submitting, you must submit a self-test report: the new session will splice the prompt into the test instruction in the AI self test , self-test the qualification of the project and submit the report, otherwise it will be directly returned
3.1 Hard threshold (One-Vote Veto)
Core principle: The code must be able to run, and it must run the content required by the problem.
3.1.1 absolutely possible
One-click startup : The deliverable must strictly support docker compose up startup. If an error is reported during the startup process (such as missing dependencies, port conflicts, configuration errors), it will be directly unqualified.
Environment isolation : It is strictly forbidden to have the situation of "can run on my local". The code must not rely on absolute paths on your local, specific global environment variables, or system libraries that are not declared in the Dockerfile.
Documentation Consistency : The startup steps in the README must be real and valid, and the validator can run without guessing or modifying the source code.
3.1.2 strict relevance
The core goal is consistent : development must strictly revolve around the business goals described in Prompt. For example, Prompt requires "implementing a customer service system that supports multiple rounds of dialogue", and only achieving "single Q & A" is considered off-topic.
Prohibit unauthorized simplification : It is strictly forbidden to reduce development difficulty by significantly reducing functionality and replacing core requirements (such as simplifying "real-time WebSocket communication" to "HTTP polling").

3.2 Delivery integrity
Core principle: Deliver product prototypes, not code snippets
3.2.1 engineered structure (0-1 completeness)
Project form : The deliverable must be a complete engineering project with a clear directory structure (such as src, config, public, tests, etc.) and code structure hierarchy .
Reject snippets : Submitting single-file code (such as a few thousand lines of main.py or index.html) or code snippets that only provide core functions is strictly prohibited. The complete configuration file (package.json, pom.xml, requirements.txt, etc.) must be included.
3.2.2 real logic implementation (reject mock spoofing)
Logical truth : Unless the Prompt explicitly requires the use of mock data (or involves expensive external APIs that cannot be called), the core business logic must be implemented truthfully.
Hardcoding is strictly prohibited : for example, the login interface cannot directly return'return "Login Success" ', and must include verification logic; the query interface cannot directly return the JSON list of hard code, and must include data query or processing logic.

3.3 Engineering and Architecture Quality
Core principles: Code must be maintainable, conform to industry norms , and meet best practices .
3.3.1 architecture layering is reasonable
Separation of responsibilities : Code structure should reflect "high cohesion, low coupling".
Backend : It is recommended to use a standard layered architecture, and it is strictly forbidden to mix database operations, business logic, and API definitions in the same function.
Front-end : Components should be reasonably split to avoid "God components" with thousands of lines.
File organization : Directory names should be semantic and easy to understand (e.g. /utils, /components, /api).
3.3.2 code cleanliness (no junk files)
Clean up redundancy: All dependency directories, cache files, and bundles must be cleaned up before submission (see Step 3 Product Attachment Structure Specification).
Configure anonymization: Ensure that the configuration file does not contain your personal key (AK/SK), intranet IP, or sensitive information.
Code quality: Remove large sections of deprecated code that have been commented out, and debug print/console.log statements.
API interface cleanliness: If there are API interfaces in the project, when receiving API interface returns, the content of the API interface must be beautified to prevent returning some unclear JSON structures. See the following example for details.
  Example：
    Untidy example: (Only reliable guessing, poor readability): Neat example: (Do pagination to ensure readability):

3.3.3 maintainability and scalability
Reject one-off code : Logic design should consider scalability and avoid a lot of Magic Numbers or deeply nested'if-else '.
3.3.4 testing standards
The project must provide a complete and executable test verification plan and related test materials as a necessary part of system acceptance. The testing goal is to verify the correctness, stability, and robustness of the system's core functions, key business logic, and exception handling mechanism.
Test verification must include both unit test and API interface functional test . Both types of tests are necessary for acceptance and must not be missing. Test requirements and examples are as follows:
3.3.4.1 unit test requirements and examples
Unit testing should cover the main functional modules, core logic processing flow, and key boundary scenarios of the system, with a focus on verifying the correctness of internal logic implementation.
Example description :
For the core business computing logic, corresponding unit test cases need to be provided to verify whether the processing results under normal input, boundary input, and illegal input conditions meet expectations.
Unit test cases should be written separately for critical state transition logic (such as task creation, execution, failure, retry, etc.) to verify the correctness of behavior in each state.
For exception handling logic, the system should verify that it can correctly return error messages and maintain stable operation by constructing exception scenarios (such as null values, out-of-range parameters, etc.).
3.3.4.2 API interface functional testing requirements and examples
API interface functional testing should cover the main interfaces provided by the system to the outside world, and verify the functional integrity and stability of the interface under different input conditions.
Example description :
For the core business interface, it is necessary to provide interface call testing to verify that the interface can return correct response results under normal request parameters.
For abnormal scenarios such as missing parameters, incorrect parameter formats, and insufficient permissions, corresponding interface test cases should be provided to verify that the error codes and error messages returned by the interface comply with the interface design specifications.
For interfaces involving data changes, the correctness of the system state or data results before and after the interface call should be verified.
3.3.4.3 test execution mode and result output requirements
The following test directory structure must be included in the project root directory as a mandatory check item during acceptance:
unit_tests/ : Used to store unit test scripts and related test resources;
API_tests/ : Used to store API interface function test scripts and related test resources.
All tests must be organized and executed uniformly through shell scripts , and the test scripts should support one-click execution and have the ability to run repeatedly.
Example description :
A unified test execution script (such as run_tests.sh ) can be provided in the project root directory, and all test cases under the unit_tests/ and API_tests/ directories can be automatically called after the script is executed.
During the test execution process, clear and readable test result information should be output in the end point or log file, including the execution status (success/failure), failure reasons, and necessary error logs for each test case.
After the test execution is completed, the summary information of the test results (such as the total number of test cases, the number of passes, and the number of failures) should be output to facilitate the acceptance personnel to quickly judge the test coverage and execution status.
3.3.4.4 Acceptance Judgment Requirements
Test cases should cover the vast majority of Functional Buttons and main business logic paths . During acceptance, the acceptance party can confirm whether the test is fully executed and the results meet expectations by executing the test script and checking the test output results.
If the unit test or API interface test is missing, the test coverage is obviously insufficient, the test script cannot be executed, or the test result output is unclear, it is considered as not meeting the acceptance requirements .
3.4 Engineering details and professionalism
Core principle: Demand yourself according to the standards of production-level code.
3.4.1 robust error handling
Elegant degradation : When the interface reports an error, it should return the standard HTTP status code and clear JSON error prompt (such as' {"code": 400, "msg": "Invalid email format"} '), and it is strictly prohibited to directly throw the original Stack Trace (stack information) or cause the service to crash without response.
Front-end fault tolerance : When the interface request fails, the UI should have the corresponding Toast prompt or default page, and cannot be white screen or no response.

3.4.2 standard logging
Effective logs : Key business processes (such as logins, payments, data changes) must have log outputs.
Log quality : Logs should contain necessary context to facilitate troubleshooting. Reject meaningless logs (such as print ("here"), console.log ("111")).

3.4.3 security and parameter validation
Input defense : validate all parameters (Body, Query, Path) passed in by the front end (null, format, length limit).
Basic security : Avoid obvious security bugs (such as concatenating SQL strings directly, storing passwords in plaintext on the front end, etc.).

3.5 Depth of Requirements Understanding
Core principle: Do it, do it right , do it well .
3.5.1 identify implicit constraints
   Business closed loop : not only to achieve the literal function, but also to think about the rationality of the business scenario. For example:
  E-commerce system: Inventory deduction cannot be negative.
  Reservation system: Reservations cannot be repeated for the same time period.
  Logic self-consistent : data flow must be logically smooth, can not appear in the foreground display success but the background is not stored.
3.5.2 reject mechanical translation
Scenario adaptation : The code implementation should conform to the user scale and usage scenarios set by Prompt, rather than blindly copying universal templates.

3.6 Aesthetics (only for front-end/full stack/mobile end questions)
Core principle: The interface should be clean, modern, and have basic interactive usability.
3.6.1 visual specification
The layout is neat : elements are aligned and spaced uniformly (Margin/Padding is reasonable), without out of memory, misalignment or garbled characters.
Harmonious color scheme : color matching in line with the main tone, appropriate contrast, not dazzling.
Modernization : It is recommended to use mainstream UI frameworks (such as Ant Design, Material UI, Tailwind CSS, Bootstrap, etc.) to improve aesthetics.
3.6.2 interaction experience
Operation feedback : button clicks should have feedback (Loading state, disabled state), mouse hover should have style changes.
Smooth process : The page jumping logic is clear, there is no dead link, and users can smoothly complete core business operations.
3.7 Unacceptable situations
The purpose of Docker delivery is to ensure "environmental consistency" and "verify low cost". If any of the following situations occur, it is considered that the deliverable is not available, directly judged as unqualified (not acceptable), and does not enter the code detail review stage .
3.7.1 Automation startup failed
Command error : When executing docker compose up in the standard Docker environment, the Build process or Run process will report an error.
Container crash : The container cannot remain running after startup (such as being trapped in a CrashLoopBackOff restart loop), or exits immediately after startup.
Private resource restriction : Dockerfile references a private mirroring repository (authentication required) or basic mirroring that cannot be accessed on the public network, resulting in pull failure.
3.7.2 rely on human intervention or implicit actions
Refuse manual configuration : Acceptance personnel need to manually create files (such as manually creating .env , manually copying config.example.js to config.js ), manually creating folders or manually importing SQL scripts to run.
Refuse interactive input : The startup script is stuck in the command line waiting for user input (such as waiting for password input, confirming y/n ), and cannot start unattended.
Relying on verbal communication : The README is not specified, and the acceptance personnel need to ask "how to run" and "what configuration is missing" through the chat tool to start.
3.7.3 environment isolation failed (can run locally)
Path Dependence : The code or configuration contains the developer's local absolute path (such as C:/Users/Admin/... or /Users/name/project/... ), causing the file to not be found in the container.
Host dependencies : Services within the container attempt to connect to a database on the host, Redis, or other services that are not declared in docker-compose.
Global environment dependencies : The code depends on libraries or tools that are not installed in the Dockerfile, but rely on the developer's local global environment installation (such as global npm packages, global python libraries).
3.7.4 document does not match actual behavior
False documentation : The verification steps (URL, API path, test account) declared in the README do not match the actual code behavior.
Port spoofing : The document claims that the service is running on port 8080 , but docker-compose exposes port 3000 without explanation.
Verification Unreachable : Following the documentation does not result in seeing the expected service interface or receiving the correct API response.
3.7.5 dependent on pollution
The project directory contains dependency directories such as node_modules/ and .venv/ , which can cause mirroring to be too large or have version conflicts during construction.

IV. Appendix: Reference Technical Standards
1. Task type classification reference
When filling out the form, please categorize tasks according to the following logic:
Task Type
Definition feature
pure_frontend
Pure web page, no real backend (Mock/LocalStorage available), focus on UI/interaction.
pure_backend
No UI or minimal UI, the core is API, data processing, and logical algorithms.
full_stack
There are both web pages and real backend services + databases.
mobile_app
Android/iOS native development, or mobile end apps like Flutter/RN.
cross_platform_app
Inter-App communication frameworks such as uni-app/Taro/Electron/Mini Program.

2. Recommended operating environment version
To ensure compatibility, it is recommended to develop and test on the following base line versions:
Front end
Node.js : 18.x LTS and above (≥ 18.16.0)
Npm 9.x or pnpm 8.x
Backend
Node.js: 18.x LTS
Python: 3.10.x
Java: 17.x LTS
External
MySQL: ≥ 8.0.x
SQLite： ≥ 3.39
PostgreSQL ≥ 14
Chrome : Version ≥ 120
