# Screensaver Application

## Project Overview

The Screensaver Application is a web-based tool designed to display various widgets on a customizable layout. This project utilizes the Gin framework for handling HTTP requests and the Open Meteo API for fetching weather data. The application supports different types of widgets, including weather forecasts, Notion calendar events, and clocks.

## Features

- **Customizable Layouts**: Users can save and load different layouts for their screensaver.
- **Weather Forecast**: Displays current, hourly, and daily weather data.
- **Notion Calendar Integration**: Fetches and displays events from a Notion database.
- **Clock Widget**: Displays the current time.
- **Responsive Design**: Uses Tailwind CSS for styling and ensuring the application is responsive.

## Project Structure

- **main.go**: The entry point of the application. It sets up the Gin router, registers routes, and starts the server.
- **apis.go**: The API endpoints of the application. It includes routes for fetching weather forecast data and Notion calendar events.
- **layout/**: Contains the code for managing and rendering layouts.
- **notion/**: Manages the integration with Notion API for fetching calendar events.
- **weather/**: Handles fetching and parsing weather forecast data from the Open Meteo API and historical data from JMA.
- **util/**: Provides utility functions and constants for the application.

## Project Configuration

The project utilizes a `project.json` file to manage scripts and dependencies essential for the development and build process. Below is a detailed explanation of the contents of this file:

### Scripts

- **cssbuild**
  - **Command**: `npx tailwindcss -i ./main.css -o ./static/tailwind.css --watch`
  - **Description**: This script uses Tailwind CSS to compile the input CSS file (`main.css`) into an output CSS file (`static/tailwind.css`). The `--watch` flag ensures that Tailwind CSS continues to monitor changes in `main.css` and updates `tailwind.css` automatically.

- **transfer-material**
  - **Command**: 
    ```bash
    rm -rf static/packages/@material/material-color-utilities && 
    mkdir -p static/packages/@material/material-color-utilities && 
    cp -r node_modules/@material/material-color-utilities/ static/packages/@material/material-color-utilities
    ```
  - **Description**: This script removes any existing material color utilities in the `static/packages/@material/material-color-utilities` directory, creates the directory if it doesn't exist, and then copies the material color utilities from `node_modules` to the static packages directory. This ensures that the required material color utilities are available in the static files of the application.

- **start**
  - **Command**: `go run *.go`
  - **Description**: This script runs the Go application. It compiles and executes all Go files in the project directory.

## Installation

1. **Clone the repository**:
   ```bash
   git clone https://github.com/kken7231/screensaver.git
   cd screensaver
   ```

2. **Install dependencies**:
   Ensure you have Go installed. Then, install the required Go packages:
   ```bash
   go get -u github.com/gin-gonic/gin
   ```

3. **Set up with npm**:
   Ensure you have npm installed. Then, configure the application:
   ```bash
   npm install
   npm run transfer-material
   npm run cssbuild
   ```


4. **Run the application**:
   ```bash
   npm run start
   ```

5. **Access the application**:
   Open your browser and navigate to `http://localhost:8080`. `http://localhost:8080?layout=<LAYOUT_NAME>` can load custom layouts in the `/layouts` directory. 

## API Endpoints

### Weather Forecast
- **Endpoint**: `/api/weatherforecast`
- **Method**: GET
- **Query Parameters**:
  - `size`: Widget size (`small` or `middlev`)
  - `location_name`: Name of the location
  - `location_latitude`: Latitude of the location
  - `location_longitude`: Longitude of the location
  - `location_histdata`: AMEDAS code for historical data (required for `middlev` size)

### Notion Calendar
- **Endpoint**: `/api/notioncalendar`
- **Method**: GET
- **Query Parameters**:
  - `size`: Widget size (`middleh`, `middlev`, or `longv`)

## Static Files

- **Design Page**: `/design` - Displays the design page.
- **Layouts Page**: `/layouts` - Displays the layouts page.
- **Tailwind CSS**: `/tailwind.css` - Serves the Tailwind CSS file.
- **JavaScript**: `/index.js` - Serves the JavaScript file.
- **Packages**: `/static/packages` - Serves static packages.

## Templates

- **Main HTML Templates**: Located in the `templates/` directory.
  - `index.tmpl`: Main template for rendering the screensaver layout.
  - `util.tmpl`: Template for utility components like horizontal lines.
- **Widget-Specific HTML Templates**: Located in the `templates/widgets` directory.
  - `clock.tmpl`: Template for rendering Clock widgets.
  - `notioncalendar.tmpl`: Template for rendering Notion calendar widgets.
  - `weatherforecast.tmpl`: Template for rendering Weather widgets.
