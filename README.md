# Examples demonstrating the Learning Tools Interoperability (LTI) library for Go

> This repository's two applications demonstrate use of our [lti](https://github.com/macewan-cs/lti) library.

## Table of Contents

- [General Information](#General-Information)
- [Technologies Used](#Technologies-Used)
- [Installation](#Installation)
- [Project Status](#Project-Status)
- [Acknowledgements](#Acknowledgements)
- [Contact](#Contact)
- [License](#License)

## General Information

The [IMS Global Learning Consortium](http://www.imsglobal.org/) developed the [Learning Tools Interoperability (LTI)](https://www.imsglobal.org/activity/learning-tools-interoperability) specification to formalize communication between learning tools and learning platforms.
The two applications in this repository, both minimal working examples of learning tools, demonstrate the use of our Go-based [lti](https://github.com/macewan-cs/lti) library.
The library partially implements version 1.3 of the specification for developers for Go-based learning tools.

One of the example, ```minimal-example.go```, provides an absolute minimal working example of the library.
It uses the library's internal nonpersistent datastore.

The other example, ```sql-example.go```, provides a slightly more complicated example.
It uses a persistent SQLite store for registrations.

Both of these examples share some code, but we have intentionally duplicated that code in the source files to make the application code is as easy to understand as possible.

## Technologies Used

- Go - version 1.16

## Installation

1. Clone this repository.
2. Create a private and public RSA keys.
```bash
openssl genrsa -out private.pem
openssl rsa -in private.pem -pubout -out public.pem
```
3. Configure the learning platform for the learning tool.

   In Moodle, go to "Site administration", "Plugins", "External tool", "Manage tools", "Configure a tool manually".
   Enter the following settings:
   - Tool settings
     - Tool name: Your choice
     - Tool URL: http://localhost:8080/launch
     - LTI version: LTI 1.3
     - Public key type: RSA key
     - Public key: Copy/paste the content of ```public.pem```
     - Initiate login URL: http://localhost:8080/login
     - Redirection URI(s): http://localhost:8080/launch
   - Services
     - IMS LTI Assignment and Grade Services: Use this service for grade sync and column management
	 - IMS LTI Names and Roles Provisioning: Use this service to retrieve members' information as per privacy settings
	 - Tool Settings: Use this service
   - Privacy:
     - Share launcher's name with tool: Always
	 - Share launcher's email with tool: Always
	 - Accept grades from the tool: Always
   Save changes.
4. In the list of external tools, display the tool configuration details.

   In Moodle, the point-form list icon within the tool's tile shows these settings.
   These settings are needed for the environment variables below.
   Use the following mappings from Moodle values to environment variable values:
   - "Platform ID" to ```REG_ISSUER```
   - "Public keyset URL" to ```REG_KEYSETURI```
   - "Access token URL" to ```REG_AUTHTOKENURI```
   - "Authentication request URL" to ```REG_AUTHLOGINURI```
5. Define the following environment variables with values appropriate for your learning platform.
```bash
export REG_ISSUER=https://platform
export REG_CLIENTID=clientid
export REG_KEYSETURI=https://platform/mod/lti/certs.php
export REG_AUTHTOKENURI=https://platform/mod/lti/token.php
export REG_AUTHLOGINURI=https://platform/mod/lti/auth.php
export REG_TARGETLINKURI=http://tool/launch
export DEP_DEPLOYMENTID=1
export KEY_PRIVATE="$(cat private.pem)"
```
6. Run the desired learning tool:
   - ```go run cmd/minimal/main.go```
   - ```go run cmd/sqlite3/main.go```
7. Attempt to launch the learning tool from the learning platform.

## Project Status

This project is under active development.

## Acknowledgements

Funding for this project was provided by the MacEwan University [Faculty of Arts and Science](https://www.macewan.ca/wcm/SchoolsFaculties/ArtsScience/AcademicPlanning/index.htm).

## Contact

Created by Ron Dyck and Nicholas M. Boers at [MacEwan University](https://www.macewan.ca/ComputerScience).

## License

This project is licensed under the MIT License.

We are actively developing the library and these examples, and we welcome all pull requests.
