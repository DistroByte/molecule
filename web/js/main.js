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
const fetchData = async (endpoint, listElement, includeFavicon = false) => {
  try {
    console.log(
      `Fetching data from ${endpoint
        .replace("/v1/urls/", "")
        .replace("/", "")}...`
    );

    const response = await fetch(endpoint);
    if (!response.ok)
      throw new Error(`Error fetching ${endpoint}: ${response.statusText}`);

    const data = await response.json();
    listElement.innerHTML = await generateListItems(data, includeFavicon);

    setupCopyableItems(listElement);
  } catch (error) {
    console.error(error);
    listElement.innerHTML = `<li>Error loading data</li>`;
  }
};

// Generate list items based on data
const generateListItems = async (data, includeFavicon) => {
  const items = await Promise.all(
    Object.entries(data).map(async ([service, value]) => {
      if (includeFavicon && value.startsWith("http")) {
        const faviconUrl = fetchFavicon(service);
        service = service.includes("-")
          ? service.slice(0, service.lastIndexOf("-"))
          : service;
        return generateListItemTemplate(
          service,
          value,
          includeFavicon,
          faviconUrl
        );
      }

      return generateListItemTemplate(service, value, includeFavicon);
    })
  );

  return items.join("");
};

// Set up copy functionality for non-URL items
const setupCopyableItems = (listElement) => {
  const copyableItems = listElement.querySelectorAll(".copyable");
  copyableItems.forEach((item) => {
    item.addEventListener("click", () => {
      const valueToCopy = item.getAttribute("data-value");
      navigator.clipboard
        .writeText(valueToCopy)
        .then(() => {
          showCopyNotification(`Copied ${valueToCopy}!`);
        })
        .catch((err) => console.error("Failed to copy text: ", err));
    });
  });
};

// Show a notification when a string is copied
const showCopyNotification = (message) => {
  const notification = document.createElement("div");
  notification.textContent = message;
  notification.style.position = "fixed";
  notification.style.bottom = "20px";
  notification.style.right = "20px";
  notification.style.backgroundColor = "#4caf50";
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
};

// Set up collapsible headers
const setupCollapsibleHeaders = () => {
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
};

// Set up refresh button functionality
const setupRefreshButton = (
  buttonId,
  endpoint,
  listElement,
  includeFavicon = false
) => {
  const button = document.getElementById(buttonId);
  if (button) {
    button.addEventListener("click", () =>
      fetchData(endpoint, listElement, includeFavicon)
    );
  }
};

// Generate favicon URL
const fetchFavicon = (service, format = "png") => {
  try {
    service = service.includes("-")
      ? service.slice(0, service.lastIndexOf("-"))
      : service;
    return `https://raw.githubusercontent.com/homarr-labs/dashboard-icons/refs/heads/main/${format}/${service}.${format}`;
  } catch (error) {
    console.error(`Error generating favicon URL for ${service}:`, error);
    return null;
  }
};

// Generate HTML for a single list item
const generateListItemTemplate = (
  service,
  value,
  includeFavicon,
  faviconUrl = null
) => {
  if (includeFavicon && value.startsWith("http")) {
    return `
      <li>
        <a href="${value}" target="_blank" style="display: flex; align-items: center; text-decoration: none; color: inherit;">
          ${
            faviconUrl
              ? `<img src="${faviconUrl}" alt="." style="width:auto; height:40px; margin-right:8px;">`
              : ""
          }
          <span>${service}</span>
        </a>
      </li>`;
  }

  return `<li class="copyable" data-value="${value}">${service}: ${value}</li>`;
};
