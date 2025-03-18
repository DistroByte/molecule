document.addEventListener("DOMContentLoaded", function () {
    fetch("/v1/urls/traefik")
        .then((response) => response.json())
        .then((data) => {
            const urlList = document.getElementById("url-list");
            for (const [key, value] of Object.entries(data)) {
                const listItem = document.createElement("li");
                const link = document.createElement("a");
                link.href = value;
                link.textContent = key;
                link.target = "_blank";
                listItem.appendChild(link);
                urlList.appendChild(listItem);
            }
        })
        .catch((error) => console.error("Error fetching URLs:", error));

    fetch("/v1/urls/hosts")
        .then((response) => response.json())
        .then((data) => {
            const hostList = document.getElementById("host-port-list");
            for (const [key, value] of Object.entries(data)) {
                const listItem = document.createElement("li");
                listItem.textContent = `${key}: ${value}`;
                hostList.appendChild(listItem);
            }
        })
        .catch((error) => console.error("Error fetching hosts:", error));

    fetch("/v1/urls/services")
        .then((response) => response.json())
        .then((data) => {
            const serviceList = document.getElementById("service-list");
            for (const [key, value] of Object.entries(data)) {
                const listItem = document.createElement("li");
                listItem.textContent = `${key}: ${value}`;
                serviceList.appendChild(listItem);
            }
        })
        .catch((error) => console.error("Error fetching services:", error));
});
