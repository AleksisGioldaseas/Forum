let imMaster = true

const bc = new BroadcastChannel("notifications");

document.addEventListener("DOMContentLoaded", function () {


    connectSSE()

    // Handler incoming messages
    bc.addEventListener("message", event => {
        imMaster = false

        if (event.data.startsWith("master:")) {
            masterLastTimestamp = event.data.slice(7)

        } else {
            const notificationCount = parseInt(event.data, 10); // Convert event.data to an integer



            const badge = document.getElementById('notification-badge');
            const button = document.getElementById('notifications-button');

            if (notificationCount == 0) {
                badge.classList.remove('active');
                badge.textContent = ""; // Remove the number when seen
                button.classList.remove('notification-active');

            } else {

                // Show the notification badge with the count
                
                if (badge) {
                    badge.classList.add('active');
                    badge.textContent = notificationCount; // Update the badge number
                } else {
                    console.log("Badge element not found");
                }

                // Add active notification style to the button and trigger background change
                if (button) {
                    button.classList.add('notification-active'); // Add the active class
                } else {
                    console.log('Button not found!');
                }

                // Optional: Add a click event to mark notification as seen
                button.addEventListener('click', () => {
                    bc.postMessage("0")
                    badge.classList.remove('active');
                    badge.textContent = ""; // Remove the number when seen
                    button.classList.remove('notification-active');
                });

            }
        }
    });

});

// master's timestamp
let masterLastTimestamp = 0;

// global variables that contro check master loop, which checks of the master exists or not using the masters heartbeat 
// (the master sends a timestamp on an interval, if that timestamp becomes out of date then the master presumably is dead)
let checkMasterLoop;
let checkMasterisLoopRunning = false;

// 1. The infinite loop with random 4.5-5 second delays
function CheckMasterLoop() {
    const minDelay = 4500; // 4.5 seconds
    const maxDelay = 5000; // 5 seconds

    const executeCheckMasterLoop = () => {

        // Your actual task logic goes here
        if (imMaster) {
            //send timestamp

            bc.postMessage("master:" + Math.floor(Date.now() / 1000).toString());
        } else {
            // If master timestamp is out of date, then connect the sse and become master

            if ((+masterLastTimestamp + 10) < +(Math.floor(Date.now() / 1000).toString())) {

                connectSSE()
            } else {

            }
        }


        // Schedule next iteration with random delay
        const delay = Math.floor(Math.random() * (maxDelay - minDelay + 1)) + minDelay;
        checkMasterLoop = setTimeout(executeCheckMasterLoop, delay);
    };

    executeCheckMasterLoop();
    checkMasterisLoopRunning = true;
}

// 2. Function to start the loop
function startLoop() {
    if (!checkMasterisLoopRunning) {

        CheckMasterLoop();
    } else {

    }
}

// 3. Function to stop the loop
function stopLoop() {
    if (checkMasterisLoopRunning) {
        clearTimeout(checkMasterLoop);
        checkMasterisLoopRunning = false;

    } else {

    }
}



function connectSSE() {
    imMaster = true
    const eventSource = new EventSource(`/ssenotifications`);

    eventSource.onmessage = (event) => {

        const notificationCount = parseInt(event.data, 10); // Convert event.data to an integer




        bc.postMessage(event.data);


        // Show the notification badge with the count
        const badge = document.getElementById('notification-badge');
        if (badge) {
            badge.classList.add('active');
            badge.textContent = notificationCount; // Update the badge number
        } else {

        }

        // Add active notification style to the button and trigger background change
        const button = document.getElementById('notifications-button');
        if (button) {
            button.classList.add('notification-active'); // Add the active class
        } else {

        }


        // Optional: Add a click event to mark notification as seen
        button.addEventListener('click', () => {
            bc.postMessage("0")
            badge.classList.remove('active');
            badge.textContent = ""; // Remove the number when seen
            button.classList.remove('notification-active');
        });

    };

    eventSource.onerror = () => {
        eventSource.close();
    };
}


startLoop()