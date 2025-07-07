document.addEventListener("DOMContentLoaded", function () {
  
    let notificationsJson = document.getElementById("notifications-data").dataset.json;

    try {
        let notifications = JSON.parse(notificationsJson);

        const activeUserRole = document.getElementById("active-user-role").dataset.value
        const activeUsername = document.getElementById("active-user-name").dataset.value

        notifications = JSON.parse(notificationsJson)
   
        refreshNotifications(notifications)
        
    } catch (error) {
        console.error("Error parsing JSON:", error);
    }

});





function refreshNotifications(notifications) {
    
    let notificationsSection = document.getElementById("notification-container");
    notificationsSection.innerHTML = ``

    if (!notifications || !Array.isArray(notifications) || notifications.length === 0) {
        notificationsSection.innerHTML = "<p>No notifications yet.</p>";
        } else {
            notifications.forEach(notif => {
                const activeUserRole = document.getElementById("active-user-role").dataset.value
                const activeUsername = document.getElementById("active-user-name").dataset.value
                let notifElement = createNotificationElement(notif, activeUserRole, activeUsername)

                notificationsSection.appendChild(notifElement);
            });
        }

    // hooknotificationEvents()
}

function appendNotifs(Notifs, activeUserRole, activeUserName) {
    let NotifsSection = document.getElementById("notification-container");

   
    Notifs.forEach(Notif => {
        n = createNotificationElement(Notif, activeUserRole, activeUserName)
        NotifsSection.appendChild(n)
    });

  
}



window.stopLoadingNotifications = false;
let lastTriggerTimeNotifications = 0;

function handleScrollNotification() {
    const now = Date.now();
    
    if (window.stopLoadingNotifications == false) {
        if ((window.innerHeight + window.scrollY) >= document.body.offsetHeight - 200) {
            if (now >= lastTriggerTimeNotifications + 1000) {
                lastTriggerTimeNotifications = now;
                fetchMoreNotifs()
                    .then(moreNotifications => {
                        const activeUserRole = document.getElementById('active-user-role').dataset.value
                        const activeUserName = document.getElementById('active-user-name').dataset.value
                        
                        
                        if (moreNotifications.length == 0) {
                            
                            window.stopLoadingNotifications = true;
                        }else{
                            if (moreNotifications){
                                appendNotifs(moreNotifications, activeUserRole, activeUserName);
                            }
                        }
                    })
                    .catch(error => {
                        
                        window.stopLoadingNotifications = true;
                        console.log("Error fetching more Notifications:", error);
                    });
            }

        }
    }

}






window.addEventListener("scroll", handleScrollNotification);