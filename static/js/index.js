import { argbFromHex, themeFromSourceColor, applyTheme } from "/static/packages/@material/material-color-utilities/index.js";

// Get the theme from a hex color
const theme = themeFromSourceColor(argbFromHex('#f82506'), [
  {
    name: "custom-1",
    value: argbFromHex("#ff0000"),
    blend: true,
  },
]);

// Check if the user has dark mode turned on
const systemDark = window.matchMedia("(prefers-color-scheme: dark)").matches;

// Apply the theme to the body by updating custom properties for material tokens
applyTheme(theme, {target: document.body, dark: systemDark});


export function updateData(widgetId, widgetType, queryString) {
  fetch(`/api/${widgetType}?${queryString}`)
      .then(response => response.json())
      .then(data => {
          // Function to convert kebab-case string to snake_case
          function toSnakeCase(str) {
              return str.replace(/([a-z])([A-Z])/g, '$1_$2').toLowerCase();
          }

          // Select all elements whose ID starts with "wgcontent-{widgetId}"
          var elements = document.querySelectorAll(`[id^="wgcontent-${widgetId}"]`);

          elements.forEach(function (element) {
              var elementId = element.id;
              var trimmedId = elementId.replace(`wgcontent-${widgetId}-`, '');
              var parts = trimmedId.split('-').map(toSnakeCase);

              // Traverse the data object using the parts array
              let value = data;
              let i = 0;
              while (i < parts.length && value !== undefined) {
                  value = value[parts[i]];
                  i++;
              }

              if (value !== undefined) {
                  if (element.classList.contains('wg-html')) {
                      element.innerHTML = value;
                  }
                  else {
                      element.innerText = value;
                  }
              }
              else {
                  element.innerText = "undefined";
              }
          });
      })
      .catch(error => console.error('Error:', error));
}

export function formatJapaneseDate(date, lang) {
  var daysOfWeek = ['日曜日', '月曜日', '火曜日', '水曜日', '木曜日', '金曜日', '土曜日'];
  var japaneseEras = [
      { name: '令和', startYear: 2019, offset: 2018 },
      { name: '平成', startYear: 1989, offset: 1988 },
      { name: '昭和', startYear: 1926, offset: 1925 },
      { name: '大正', startYear: 1912, offset: 1911 },
      { name: '明治', startYear: 1868, offset: 1867 }
  ];

  if(lang === "EN") {
    daysOfWeek = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];
    japaneseEras = [
        { name: 'Reiwa', startYear: 2019, offset: 2018 },
        { name: 'Heisei', startYear: 1989, offset: 1988 },
        { name: 'Showa', startYear: 1926, offset: 1925 },
        { name: 'Taisho', startYear: 1912, offset: 1911 },
        { name: 'Meiji', startYear: 1868, offset: 1867 }
    ];
  }

  const year = date.getFullYear();
  const month = date.getMonth() + 1; // Months are zero-based in JavaScript
  const day = date.getDate();
  const dayOfWeek = daysOfWeek[date.getDay()];

  // Find the current era
  const currentEra = japaneseEras.find(era => year >= era.startYear);
  const eraYear = year - currentEra.offset;

  return `${year}(${currentEra.name} ${eraYear})/${month}/${day} ${dayOfWeek}`;
}

export function formatTime(date) {
  const hours = String(date.getHours()).padStart(2, '0');
  const minutes = String(date.getMinutes()).padStart(2, '0');
  const seconds = String(date.getSeconds()).padStart(2, '0');

  return `${hours}:${minutes}:${seconds}`;
}