import amqp, { Connection, Channel } from "amqplib";
let connection: Connection | null = null;
let channel: Channel | null = null;
async function connect_rabbitmq() {
  if (connection && channel) {
    return { connection, channel };
  }
  try {
    connection = await amqp.connect("amqp://localhost");
    channel = await connection.createChannel();

    console.log("Connected to RabbitMQ");

    return { connection, channel };
  } catch (error) {
    console.error("Failed to connect to RabbitMQ:", error);
    throw error;
  }
}
export default connect_rabbitmq;
