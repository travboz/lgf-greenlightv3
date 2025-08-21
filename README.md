# Let's Go Further - Go Web Development  

![Gopher playing piano - Greenlight Movie Database](https://raw.githubusercontent.com/egonelbre/gophers/63b1f5a9f334f9e23735c6e09ac003479ffe5df5/vector/arts/upright.svg)

This repository contains my code and exercises while following *Let's Go Further* by Alex Edwards. The project focuses on building a production-ready Go web application, covering key concepts like authentication, middleware, database management, background tasks, and deployment.  

## Features  
- RESTful JSON API
- Authentication & authorization
- Middleware implementation
- Database management with migrations
- Background task processing
- Graceful shutdown & observability

## Getting Started  

### Prerequisites  
- Go (latest stable version)  
- Docker
- Git  

### Installation  
1. Clone the repository:  
    ```bash
    git clone https://github.com/travboz/lgf-greenlightv3.git
    cd lgf-greenlightv3
    ```
2. Install dependencies:
   ```bash 
   go mod tidy
   ```
3. Configure environment variables

### Running the application
1. Set up the dependencies for the app:
    ```bash
    make setup
    ```
2. Then, run the server by navigating to the base directory and using: 
    ```bash
    make run
    ```
To run the applicaiton using configured defaults.
3. Navigate to `http://localhost:4000/v1/healthcheck` to confirm API status


## Contributing
Feel free to fork and submit PRs!

## License:
`MIT`

This should work for GitHub! Let me know if you need any tweaks. 

## Image
Image by [Egon Elbre](https://github.com/egonelbre), used under CC0-1.0 license.