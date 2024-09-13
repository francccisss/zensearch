import polling from "../utils/job_polling.js";
import client_websocket from "./websocket.js";

//polling.init(
//  () => {
//    console.log("Polling search request.");
//  },
//  (webpages) => {
//    if (webpages == undefined) {
//      console.log("No jobs to poll.");
//      return;
//    }
//    console.log("Polling done for search query.");
//  },
//);
