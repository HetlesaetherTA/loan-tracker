// handle header unauthorized
const originalFetch = window.fetch;


window.fetch = async (...args) => {
  const response = await originalFetch(...args);

  if (response.status === 401) {
    const data = await response.json();
    console.log("returning to", data.redirect_to);
    window.location.replace(data.redirect_to);
    return new Promise(() => { });
  }

  return response;
};
