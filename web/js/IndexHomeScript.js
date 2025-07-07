
let selectedCategories = [];

// Document ready event listener to initialize all interactions
document.addEventListener('DOMContentLoaded', () => {
    initializeButtonGroups();      // Initialize button groups (filtering and sorting)
    initializeCategorySelection(); // Initialize category selection
    initializeSidebarToggle();     // Initialize sidebar toggle behavior
});

// Function to initialize the filtering and sorting button groups
function initializeButtonGroups() {
    // Add event listeners to both button groups
    document.querySelectorAll('.button-group').forEach(group => {
        group.querySelectorAll('button').forEach(button => {
            // Set up click listener for each button in the group
            button.addEventListener('click', (event) => {
                event.preventDefault(); // Prevent default behavior if necessary

                // Handle the active state toggle for the clicked button
                handleActiveState(group, button);

                // Call the existing searchButtonHandler function
                searchButtonHandler();
            });
        });
    });

    // Set default active buttons when the page loads
    let filterButton = document.querySelector('#filtering button[data-value="all"]');
    let sortButton = document.querySelector('#sorting button[data-value="hot"]');
    if (filterButton){
        filterButton.classList.add('active');
    }
    if (sortButton){
        sortButton.classList.add('active');
    }
    

}

// Function to manage the active state of the buttons in each group
function handleActiveState(group, clickedButton) {
    // Remove active class from all buttons in the group
    group.querySelectorAll('button').forEach(button => {
        button.classList.remove('active');
    });

    // Add the active class to the clicked button
    clickedButton.classList.add('active');
}

// Function to initialize category selection
function initializeCategorySelection() {

    const categoryOptions = document.querySelectorAll('.category-option');
    const selectedCategoriesContainer = document.getElementById('selected-categories-container');

    // Add click event listeners to each category option
    categoryOptions.forEach(option => {
        option.addEventListener('click', () => {
            const category = option.textContent.trim(); // Get the category text

            // Prevent adding more than 5 categories
            if (selectedCategories.length >= 5 && !selectedCategories.includes(category)) {
                showErrorPopup("You can only select up to 5 categories!"); // Show the error popup
                return; // Prevent adding the category
            }

            // Toggle the 'selected' class on the clicked category
            option.classList.toggle('selected');

            // If the category is already in the selected list, remove it
            if (selectedCategories.includes(category)) {
                selectedCategories = selectedCategories.filter(cat => cat !== category);
            } else {
                // Add the category to the selected list
                selectedCategories.push(category);
            }

            updateSelectedCategories();  // Update the selected categories display
            filterPostsBySelectedCategories(); // Filter posts based on selected categories
        });
    });

    // Function to update the "Currently Selected" categories display
    function updateSelectedCategories() {
        selectedCategoriesContainer.innerHTML = ''; // Clear current content

        selectedCategories.forEach(category => {
            const categoryBox = document.createElement('div');
            categoryBox.classList.add('selected-category-box');
            categoryBox.textContent = category;
            selectedCategoriesContainer.appendChild(categoryBox);
        });

        // Show or hide the "Currently Selected" container based on selections
        selectedCategoriesContainer.style.display = selectedCategories.length > 0 ? 'block' : 'none';
    }

    // Function to filter posts based on selected categories
    function filterPostsBySelectedCategories() {
        searchButtonHandler();
    }
}

// Function to initialize the sidebar toggle behavior
function initializeSidebarToggle() {
    document.getElementById("sidebar-toggle-btn").addEventListener("click", function () {
        const panel = document.getElementById("category-panel");

        // Toggle the sliding panel
        if (panel.style.display === "none" || panel.style.display === "") {
            panel.style.display = "block";
            setTimeout(() => {
                panel.style.transform = "translateX(0)";  // Slide panel into view from the left
            }, 10);  // Wait a moment to ensure the display change is registered before sliding
        } else {
            panel.style.transform = "translateX(-100%)";  // Slide panel off-screen to the left
            setTimeout(() => {
                panel.style.display = "none";  // Hide the panel after it has slid out
            }, 300);  // Match the duration of the slide transition
        }
    });
}

// JSON Parsing and Batch Rendering
let allPosts = [];
let currentIndex = 0;
const postsPerBatch = 10;


window.stopLoading = false;
let lastTriggerTime = 0;

function handleScroll() {
    const now = Date.now();
    if (window.stopLoading == false) {
        if ((window.innerHeight + window.scrollY) >= document.body.offsetHeight - 200) {
            if (now >= lastTriggerTime + 1000) {
                lastTriggerTime = now;
                fetchMorePosts()
                    .then(morePosts => {
                        const activeUserRole = document.getElementById('active-user-role').dataset.value
                         const activeUserName = document.getElementById('active-user-name').dataset.value
                        appendPosts(morePosts, activeUserRole, activeUserName);
                        if (morePosts.length == 0) {
                            window.stopLoading = true;
                        }
                    })
                    .catch(error => {
                        window.stopLoading = true;
                        console.error("Error fetching more posts:", error);
                    });
            }

        }
    }

}





window.addEventListener("scroll", handleScroll);