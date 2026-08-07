console.log("test")

// format dates
document.querySelectorAll(".date-field").forEach(e => {
  const rawDate = e.dataset.rawDate;
  if (!rawDate) return;

  const date = new Date(rawDate)

  let include = {
    year: "numeric",
    month: "short",
  }

  if (e.dataset.precision === "month") {
    e.textContent = date.toLocaleDateString(undefined, include)
    return
  }

  include.day = "numeric"

  if (e.dataset.precision === "day") {
    e.textContent = date.toLocaleDateString(undefined, include);
    return
  };

  include.hour = "numeric"
  include.minute = "numeric"

  e.textContent = date.toLocaleDateString(undefined, include);
});
