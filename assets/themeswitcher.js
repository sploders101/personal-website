const themes = [
  "coltons-choice",
  "oceanblue-decor",
  "darkred-decor",
  "pink-decor",
  "seagreen-decor",
  "brown-decor",
  "electricblue-decor",
];
window.addEventListener("keypress", (key) => {
  if (key.key !== "t") return;
  const currentTheme = themes.find((theme) => document.body.classList.contains(theme));
  const themeIndex = themes.indexOf(currentTheme);
  const newTheme = themes[(themeIndex + 1) % themes.length];
  document.body.classList.remove(currentTheme);
  document.body.classList.add(newTheme);
});
