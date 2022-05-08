# Qualtrics Post Appointment Survey Response Visualization.
Welcome! You can find the code for the Qualtrics Post Appointment Survey Dashboard here. The project is hosted at the follwoing website:
https://qualtrics-vis.herokuapp.com/

It may take 30 ~ 60 seconds when accessing it for the first time in a while. This is because Heroku puts the application to sleep after it has been inactive for more than 30minutes. Read more about it [here](https://devcenter.heroku.com/articles/free-dyno-hours).

## Strcture of the program:
There are two main parts to this application: one part that displays and visualizes the data; the other part that fetches and processes the data. The code in this repository is responsible for the former, while the latter code can be found [here](https://github.com/kshuta/qualtrics_vis_scripts). 

This program uses Go for backend, and a library called Chart.js for the frontend to display the data. Not gonna lie, the code is not as organized as I wanted the to be, but please read through them to see how I display data. Most of the code should be in server.go.
I also have a test file, but that literally does nothing, so you're better off writing your own tests.

