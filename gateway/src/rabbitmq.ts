import amqp, { Connection, Channel } from "amqplib";
let connection: Connection | null = null;
async function connect_rabbitmq(): Promise<Connection | null> {
  if (connection) {
    return connection;
  }
  try {
    connection = await amqp.connect("amqp://localhost");

    console.log("Connected to RabbitMQ");

    return connection;
  } catch (error) {
    console.error("Failed to connect to RabbitMQ:", error);
    throw error;
  }
}
export default connect_rabbitmq;
