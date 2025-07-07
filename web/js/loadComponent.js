document.addEventListener("DOMContentLoaded", function () {

    // Load Header
    const headerModule = document.getElementById("header-module");
    const headerContainer = document.getElementById("header-container");
    headerContainer.appendChild(headerModule) // This RELOCATES the element
    
    // Right block
    const rightBlock = document.getElementById("right-block");
    const rightBlockContainer = document.getElementById("right-block-container");
    rightBlockContainer.appendChild(rightBlock)

    // Load Modals
    const modalsModule = document.getElementById("modals-module");
    const modalsContainer = document.getElementById("modals-container");
    modalsContainer.appendChild(modalsModule)
    


    initModals(); // Initialize modal functions AFTER loading

    const contentLoadedEvent = new CustomEvent("dynamicContentLoaded");
    document.dispatchEvent(contentLoadedEvent);

    function initModals() {
        const registerModal = document.getElementById("register-modal");
        const loginModal = document.getElementById("login-modal");
        const backdrop = document.getElementById("modal-backdrop");
        const contentContainer = document.getElementById("content-container");

        const openRegisterBtns = document.querySelectorAll("#open-register");
        const openLoginBtns = document.querySelectorAll("#auth-btn");

        const closeRegisterBtn = document.querySelector(".close");
        const closeLoginBtn = document.querySelector(".close-login");

        const loginSubmitBtn = document.getElementById("sendButtonLogin");
        const registerSubmitBtn = document.querySelector(".submit-btn");

        function openModal(modal) {
            modal.style.display = "block";
            backdrop.style.display = "block";
            contentContainer.classList.add("blurred");
            document.body.classList.add("modal-open");
            // showStep(1); // < ERROR "showStep" IS UNDEFINED! (Alex: I saw it on the console when checking stuff on the site) Ensure first step is always shown when opening the register modal
        }

        function closeModal(modal) {
            modal.style.display = "none";
            backdrop.style.display = "none";
            contentContainer.classList.remove("blurred");
            document.body.classList.remove("modal-open");
        }

        openRegisterBtns.forEach(btn => {
            btn.addEventListener("click", () => openModal(registerModal));
        });

        openLoginBtns.forEach(btn => {
            btn.addEventListener("click", () => openModal(loginModal));
        });

        if (closeRegisterBtn) closeRegisterBtn.addEventListener("click", () => closeModal(registerModal));
        if (closeLoginBtn) closeLoginBtn.addEventListener("click", () => closeModal(loginModal));

        // Close modal if clicking outside (on backdrop)
        window.addEventListener("click", function (event) {
            if (event.target === backdrop) {
                closeModal(registerModal);
                closeModal(loginModal);
            }
        });

        // Handle form submissions
        if (registerSubmitBtn) {
            registerSubmitBtn.addEventListener("click", function () {
                const username = document.getElementById("username").value;
                const password = document.getElementById("password").value;

                if (username && password) {
                    closeModal(registerModal);
                }
            });
        }

        if (loginSubmitBtn) {
            loginSubmitBtn.addEventListener("click", function () {
                const username = document.getElementById("loginUsername").value;
                const password = document.getElementById("loginPassword").value;

                if (username && password) {
                    closeModal(loginModal);
                }
            });
        }

    }
    

});


