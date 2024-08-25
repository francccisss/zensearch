import express, { Request, Response, NextFunction } from "express";

const app = express();
const PORT = 8080;

app.listen(PORT, () => {
  console.log("Listening to Port:", PORT);
});

app.get("/", (req: Request, res: Response) => {
  res.send({ Message: "Initialized frontend" });
});
