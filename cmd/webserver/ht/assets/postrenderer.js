window.addEventListener("load", () => {
  const timeElements = document.querySelectorAll("[js-localtime]");
  timeElements.forEach((element) => {
    const timestamp = element.getAttribute("js-localtime");
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
