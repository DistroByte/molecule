document.addEventListener("DOMContentLoaded", () => {
  const urlList = document.getElementById("url-list");
  const hostPortList = document.getElementById("host-port-list");
  const serviceList = document.getElementById("service-list");

  const fetchFavicon = (service, format = "png") => {
    try {
      // slice off the final dash and everything after it if one exists
      if (service.includes("-")) {
        service = service.slice(0, service.lastIndexOf("-"));
      }
      const faviconUrl = `https://raw.githubusercontent.com/homarr-labs/dashboard-icons/refs/heads/main/${format}/${service}.${format}`;
      // check if the favicon exists
      return faviconUrl; // Return the constructed favicon URL
    } catch (error) {
      console.log(`Error generating favicon URL for ${service}:`, error);
      return null; // Return null if an error occurs
    }
  };

  const fetchData = async (endpoint, listElement, includeFavicon = false) => {
    try {
      const response = await fetch(endpoint);
      if (!response.ok) {
        throw new Error(`Error fetching ${endpoint}: ${response.statusText}`);
      }
      const data = await response.json();

      listElement.innerHTML = await Promise.all(
        Object.entries(data).map(async ([service, value]) => {
          if (includeFavicon && value.startsWith("http")) {
            const faviconUrl = fetchFavicon(service);
            if (service.includes("-") && includeFavicon) {
              service = service.slice(0, service.lastIndexOf("-"));
            }

            return `<li>
                <a href="${value}" target="_blank" style="display: flex; align-items: center; text-decoration: none; color: inherit;">
                  ${
                    faviconUrl
                      ? `<img src="${faviconUrl}" alt="." style="width:40px; height:40px; margin-right:8px;">`
                      : ""
                  }
                  <span>${service}</span>
                </a>
              </li>`;
          }

          return `<li>${service}: ${value}</li>`;
        })
      ).then((items) => items.join(""));
    } catch (error) {
      console.error(error);
      listElement.innerHTML = `<li>Error loading data</li>`;
    }
  };

  fetchData("/v1/urls/traefik", urlList, true);
  fetchData("/v1/urls/hosts", hostPortList);
  fetchData("/v1/urls/services", serviceList);

  const collapsibleHeaders = document.querySelectorAll(".collapsible-header");

  collapsibleHeaders.forEach((header) => {
    header.addEventListener("click", () => {
      const list = header.nextElementSibling; // Get the associated list
      const caret = header.querySelector(".caret"); // Get the caret inside the header
      if (list && list.classList.contains("collapsible")) {
        if (list.classList.contains("expanded")) {
          // Collapse
          list.style.maxHeight = null;
        } else {
          // Expand
          list.style.maxHeight = list.scrollHeight + "px";
        }
        list.classList.toggle("expanded"); // Toggle the expanded class on the list
        if (caret) {
          caret.classList.toggle("rotate"); // Toggle the rotate class on the caret
        }
      }
    });
  });
});
