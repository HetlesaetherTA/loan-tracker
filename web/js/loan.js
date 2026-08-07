// Accrued Interest 
function calculateAccruedInterest() {
  const overlay = document.getElementById("overlay");

  const rawCalculatedAt = overlay.dataset.interestCalculatedAt;
  const principal = parseFloat(overlay.dataset.principal);
  const yearlyInterestPercent = parseFloat(overlay.dataset.yearlyInterest);
  const currency = overlay.dataset.currency;

  console.log(rawCalculatedAt, principal, yearlyInterestPercent)
  if (!rawCalculatedAt || isNaN(principal) || isNaN(yearlyInterestPercent)) {
    return 0;
  }

  const calculatedAtDate = new Date(rawCalculatedAt);
  const now = new Date();

  const diffTime = now - calculatedAtDate;
  if (diffTime <= 0) return 0;

  const diffYears = diffTime / (1000 * 60 * 60 * 24 * 365.25);

  const rate = yearlyInterestPercent / 100;
  const accruedInterest = principal * rate * diffYears;

  return currency + " " + accruedInterest.toFixed(2);
}

const accruedInterest = calculateAccruedInterest();
document.getElementById("accrued-interest").textContent = accruedInterest


// Handle overlay
function toggleOverlay() {
  const overlay = document.getElementById("overlay")


  if (overlay.classList.contains("hidden")) {
    overlay.classList.remove("hidden")
  } else {
    overlay.classList.add("hidden")
  }
}

// click outside overlay window
document.getElementById("overlay").addEventListener("click", (event) => {
  const infoWindow = document.querySelector(".info-window");

  if (!infoWindow.contains(event.target)) {
    toggleOverlay();
  }
});

// click close button
document.querySelector(".close-btn").addEventListener("click", toggleOverlay)

// click (i) to open overlay
document.getElementById("interest-details").addEventListener("click", toggleOverlay)

// Color Transactions Container Items
document.querySelectorAll(".transaction-item").forEach(item => {
  const amountSpan = item.querySelector(".tx-amount");
  const amountText = amountSpan.textContent;

  if (amountText.includes("-")) {
    amountSpan.style.backgroundColor = "#e6f4ea";
    amountSpan.style.color = "#137333";
    amountSpan.style.padding = "2px 6px";
    amountSpan.style.borderRadius = "4px";
  } else {
    amountSpan.style.backgroundColor = "#fce8e6";
    amountSpan.style.color = "#c5221f";
    amountSpan.style.padding = "2px 6px";
    amountSpan.style.borderRadius = "4px";
  }
});
