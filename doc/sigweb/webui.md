# SIGWeb UI

This document details the technical details and functionalities of the SIG configuration interface.

## Architecture

Angular is used as a frontend framework (more info in `go/sigmgmt/sigweb/README`) with Angular
Material for layouting. Please refer to the Angular documentation `https://angular.io/docs` for more
information about the testing frameworks (karma, protractor). An introduction to reactive
programming which is used in this app can be found on `http://reactivex.io/rxjs/`. Observables are
used for async HTTP requests and allow components (views) to subscribe to the result.

The app consists of the following main components:

*   Login
*   Sites
*   Contact
*   Config: Loads and displays markdown

This can be easily extended when more functionality is required.

### API

The Go API is called for all user specific information. The app has the following services which
handle API communication:

*   ApiService: Provides a method for every call which is made to the api. E.g. `getSites()`
*   UserService: Keeps track if a user is currently authenticated against the API (if JWT is
    present).

Furthermore HttpInterceptors are used to set the JWT (@auth0/angular-jwt) and request headers and
for logging requests/responses.

The backend is providing a REST API using `gorilla/mux` as router and `negroni` as a middleware for
request handling. The `dgrijalva/jwt-go` library is used for token handling. Sqlite3 is used as a
database.

### JWT

JSON Web Tokens (jwt.io) are used to authenticate a user. The token contains the username, its
feature level and expiration dates of the token.

## Functionality

The main functionality of the WebUI is to add and edit sites and their corresponding configuration
values. All this functionality can be found in the `sites` component. There exist the following
subcomponents:

*   Site Configuration: Edit site details
*   Management: Request to generate and push config file
*   Path Selectors: Add and remove path predicates
*   AS: Add ASes

For every AS the following sub-components exist:

*   Networks: Edit remote network CIDR
*   Remote SIG: Edit remote SIG details
*   Sessions: Define sessions, based on path predicates
*   Traffic Policies: Edit traffic policies
