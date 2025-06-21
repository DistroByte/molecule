document.addEventListener("DOMContentLoaded", () => {
  const urlList = document.getElementById("url-list");
  const hostPortList = document.getElementById("host-port-list");
  const serviceList = document.getElementById("service-list");

  // Initial data fetch
  fetchData("/v1/urls/traefik", urlList, true);
  fetchData("/v1/urls/hosts", hostPortList);
  fetchData("/v1/urls/services", serviceList);

  // Set up collapsible headers
  setupCollapsibleHeaders();

  // Set up refresh buttons
  setupRefreshButton("refresh-urls-button", "/v1/urls/traefik", urlList, true);
  setupRefreshButton("refresh-hosts-button", "/v1/urls/hosts", hostPortList);
  setupRefreshButton(
    "refresh-services-button",
    "/v1/urls/services",
    serviceList
  );
});

// Function to fetch data and populate the list
async function fetchData(endpoint, listElement, includeFavicon = false) {
  try {
    console.info(
      `Fetching data from ${endpoint
        .replace("/v1/urls/", "")
        .replace("/", "")}...`
    );

    const response = await fetch(endpoint);
    if (!response.ok)
      throw new Error(`Error fetching ${endpoint}: ${response.statusText}`);

    const data = await response.json();
    console.debug(`Data fetched from ${endpoint}:`, data);
    listElement.innerHTML = await generateListItems(data, includeFavicon);

    try {
      setupCopyableItems(listElement);
    } catch (error) {
      console.error("Error setting up copyable items:", error);
      listElement.innerHTML = `<li>Error setting up copyable items</li>`;
    }

    return response;
  } catch (error) {
    console.error(error);
    listElement.innerHTML = `<li>Error loading data</li>`;
    return null;
  }
}

// Generate list items based on data
async function generateListItems(data, includeFavicon) {
  const items = await Promise.all(
    data.map(async (entry) => {
      if (includeFavicon && entry.url.startsWith("http")) {
        const faviconUrl = fetchFavicon(entry.service);
        entry.service = entry.service.includes("-")
          ? entry.service.slice(0, entry.service.lastIndexOf("-"))
          : entry.service;
        return generateListItemTemplate(
          entry.service,
          entry.url,
          entry.fetched,
          includeFavicon,
          faviconUrl
        );
      }

      return generateListItemTemplate(
        entry.service,
        entry.url,
        entry.fetched,
        includeFavicon
      );
    })
  );

  return items.join("");
}

// Set up copy functionality for non-URL items
function setupCopyableItems(listElement) {
  const copyableItems = listElement.querySelectorAll(".copyable");
  copyableItems.forEach((item) => {
    item.addEventListener("click", () => {
      const valueToCopy = item.getAttribute("data-value");
      navigator.clipboard
        .writeText(valueToCopy)
        .then(() => {
          showCopyNotification(`Copied ${valueToCopy}!`);
        })
        .catch((err) => {
          console.error("Failed to copy text:", err);
          showCopyNotification("Failed to copy text.", true);
        });
    });
  });
}

// Show a notification when a string is copied
function showCopyNotification(message, isError = false) {
  const notification = document.createElement("div");
  notification.textContent = message;
  notification.style.position = "fixed";
  notification.style.bottom = "20px";
  notification.style.right = "20px";
  notification.style.backgroundColor = isError ? "#f44336" : "#4caf50"; // Red for errors, green for success
  notification.style.color = "#fff";
  notification.style.padding = "10px 20px";
  notification.style.borderRadius = "5px";
  notification.style.boxShadow = "0 2px 5px rgba(0, 0, 0, 0.2)";
  notification.style.zIndex = "1000";
  notification.style.fontSize = "14px";

  document.body.appendChild(notification);

  // Remove the notification after 3 seconds
  setTimeout(() => {
    notification.remove();
  }, 3000);
}

// Set up collapsible headers
function setupCollapsibleHeaders() {
  const collapsibleHeaders = document.querySelectorAll(".collapsible-header");
  collapsibleHeaders.forEach((header) => {
    header.addEventListener("click", () => {
      const parentContainer = header.parentElement;
      const list = parentContainer.nextElementSibling;
      const caret = header.querySelector(".caret");

      if (list && list.classList.contains("collapsible")) {
        list.style.maxHeight = list.classList.contains("expanded")
          ? null
          : `${list.scrollHeight}px`;
        list.classList.toggle("expanded");
        if (caret) caret.classList.toggle("rotate");
      }
    });
  });
}

// Set up refresh button functionality
function setupRefreshButton(
  buttonId,
  endpoint,
  listElement,
  includeFavicon = false
) {
  const button = document.getElementById(buttonId);
  if (button) {
    button.addEventListener("click", () =>
      fetchData(endpoint, listElement, includeFavicon).then(() =>
        showCopyNotification(
          `Refreshed ${endpoint
            .replace("/v1/urls/", "")
            .replace("/", "")} list!`
        )
      )
    );
  }
}

// Generate favicon URL
function fetchFavicon(service, format = "png") {
  try {
    service = service.includes("-")
      ? service.slice(0, service.lastIndexOf("-"))
      : service;
    return `https://raw.githubusercontent.com/homarr-labs/dashboard-icons/refs/heads/main/${format}/${service}.${format}`;
  } catch (error) {
    console.error(`Error generating favicon URL for ${service}:`, error);
    return null;
  }
}

// Generate HTML for a single list item
function generateListItemTemplate(
  service,
  url,
  fetched,
  includeFavicon,
  faviconUrl = null
) {
  try {
    if (includeFavicon && url.startsWith("http")) {
      return `
      <li>
        <a href="${url}" target="_blank" style="display: flex; align-items: center; text-decoration: none; color: inherit;">
          ${
            faviconUrl
              ? `<img src="${faviconUrl}" alt="&ZeroWidthSpace;" style="width:auto; height:40px; margin-right:10px;">`
              : ""
          }
          <span>${service}</span>
        </a>
        ${
          fetched
            ? `<button class="restart-button" data-service="${service}" style="margin-left: 10px;">R</button>`
            : ""
        }
      </li>`;
    }

    return `
    <li class="copyable" data-value="${url}">
      ${service}: ${url}
      <button class="restart-button" data-service="${service}" style="margin-left: 10px;">R</button>
    </li>`;
  } catch (error) {
    console.error(`Error generating list item for ${service}:`, error);
    return `<li>Error generating item for ${service}</li>`;
  }
}

document.addEventListener("click", (event) => {
  if (event.target.classList.contains("restart-button")) {
    const service = event.target.getAttribute("data-service");
    restartService(service);
  }
});

function restartService(service) {
  console.log(`Restarting service: ${service}`);

  // Show the authentication modal
  const authModal = document.getElementById("auth-modal");
  const authForm = document.getElementById("auth-form");
  const authCancel = document.getElementById("auth-cancel");

  authModal.style.display = "flex";

  // Handle form submission
  const handleAuthSubmit = (event) => {
    event.preventDefault();

    const apiKey = document.getElementById("auth-apikey").value;
    document.getElementById("auth-apikey").value = ""; // Clear the input field

    if (!apiKey) {
      alert("An API key is required to restart the service.");
      return;
    }

    // Hide the modal
    authModal.style.display = "none";

    // Remove event listeners to prevent duplicate submissions
    authForm.removeEventListener("submit", handleAuthSubmit);
    authCancel.removeEventListener("click", handleAuthCancel);

    // Make the fetch request
    fetch(`/v1/services/${service}/alloc-restart`, {
      method: "POST",
      headers: {
        "X-API-KEY": apiKey,
      },
    })
      .then((response) => {
        if (!response.ok) {
          throw new Error(`Failed to restart service: ${response.statusText}`);
        }
        showRestartNotification(`Service ${service} restarted successfully!`);
      })
      .catch((error) => {
        console.error(`Error restarting service ${service}:`, error);
        showRestartNotification(`Failed to restart service ${service}.`, true);
      });
  };

  // Handle cancel button click
  const handleAuthCancel = () => {
    authModal.style.display = "none";
    authForm.removeEventListener("submit", handleAuthSubmit);
    authCancel.removeEventListener("click", handleAuthCancel);
  };

  authForm.addEventListener("submit", handleAuthSubmit);
  authCancel.addEventListener("click", handleAuthCancel);
}

// Function to show a restart notification
function showRestartNotification(message, isError = false) {
  const notification = document.createElement("div");
  notification.textContent = message;
  notification.style.position = "fixed";
  notification.style.bottom = "20px";
  notification.style.right = "20px";
  notification.style.backgroundColor = isError ? "#f44336" : "#4caf50"; // Red for errors, green for success
  notification.style.color = "#fff";
  notification.style.padding = "10px 20px";
  notification.style.borderRadius = "5px";
  notification.style.boxShadow = "0 2px 5px rgba(0, 0, 0, 0.2)";
  notification.style.zIndex = "1000";
  notification.style.fontSize = "14px";

  document.body.appendChild(notification);

  // Remove the notification after 3 seconds
  setTimeout(() => {
    notification.remove();
  }, 3000);
}

document.addEventListener("DOMContentLoaded", () => {
  const showApiKeyButton = document.getElementById("show-apikey");
  const apiKeyInput = document.getElementById("auth-apikey");

  showApiKeyButton.addEventListener("click", () => {
    if (apiKeyInput.type === "password") {
      apiKeyInput.type = "text";
      showApiKeyButton.textContent = "Hide"; // Change icon to "hide"
    } else {
      apiKeyInput.type = "password";
      showApiKeyButton.textContent = "Show"; // Change icon to "show"
    }
  });
});
