/*

# postrenderer.js

This project generally tries to avoid JavaScript to make the site as accessible
& light as possible (it even runs in lynx!). However, there are some
quality-of-life improvements that are not possible with HTML alone. This file
adds miscellaneous scripts to add these quality-of-life improvements as
optional add-ons.


## Date Localization

The server should always format dates in its local timezone. However, it can
also provide a UNIX timestamp in a `data-unixtime` attribute to have it
converted to the user's local timezone after page load.

*/

window.addEventListener("load", () => {
  const timeElements = document.querySelectorAll("[data-unixtime]");
  timeElements.forEach((element) => {
    const timestamp = element.getAttribute("data-unixtime");
    if (timestamp === "null") return;
    const date = new Date(Number(timestamp));
    element.innerText = date
      .toLocaleString("en-US", {
        month: "2-digit",
        day: "2-digit",
        year: "2-digit",
        hour: "2-digit",
        minute: "2-digit",
        hour12: false,
      })
      .replace(",", "");
  });
});
