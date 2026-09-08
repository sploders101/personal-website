const themes = [
  "coltons-choice",
  "oceanblue-decor",
  "darkred-decor",
  "pink-decor",
  "seagreen-decor",
  "brown-decor",
  "electricblue-decor",
];
window.addEventListener("load", () => {
  const themeSwitch = document.getElementById("themeSwitcher");
  if (themeSwitch !== undefined) {
    themeSwitch.addEventListener("click", () => {
      const currentTheme = themes.find((theme) => document.body.classList.contains(theme));
      const themeIndex = themes.indexOf(currentTheme);
      const newTheme = themes[(themeIndex + 1) % themes.length];
      document.body.classList.remove(currentTheme);
      document.body.classList.add(newTheme);
    });
  }
}, { once: true });
