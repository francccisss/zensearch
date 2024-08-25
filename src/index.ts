import express, { Request, Response, NextFunction } from "express";
import path from "path";

const app = express();
const PORT = 8080;

app.listen(PORT, () => {
  console.log("Listening to Port:", PORT);
});

app.use(express.static(path.join(__dirname, "public")));

app.use(
  express.static(path.join(__dirname, "public/scripts")),
  (req: Request, res: Response) => {
    console.log(req.originalUrl);
    if (req.path.endsWith(".js")) {
      res.setHeader("Content-Type", "application/javascript");
    }
  },
);
app.get("/", (req: Request, res: Response) => {
  res.sendFile(path.join(__dirname, "public", "index.html"));
});
