import utils from "./utils";

(function main([_, , ...query_params]: Array<string>) {
  const user_query = query_params[0];
  console.log(user_query);
  const webpages = utils.yaml_loader<{ docs: Array<string> }>(
    "src/utils/webpage_database.yaml",
  );
  console.log(webpages.docs);
})(process.argv);
