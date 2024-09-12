import polling from "../utils/job_polling.js";

polling.init(
  () => {
    console.log("Polling search request.");
  },
  () => {
    console.log("Polling done for search query.");
  },
);
