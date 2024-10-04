import os
import sys
import logging
import uuid
from pyspark.sql import SparkSession
from pyspark.sql.functions import window, count, current_timestamp, col, from_json, collect_list, struct, to_json
from pyspark.sql.types import *
from pyspark.sql.types import StructType, StructField, StringType
from dotenv import load_dotenv

# Set up logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

load_dotenv()

# app conf
app_name = os.getenv('APP_NAME')
app_mode = os.getenv('APP_MODE')
PROCEESSING_TIME = os.getenv('PROCEESSING_TIME')

# Kafka configurations
kafka_boostrap_server = os.getenv('KAFKA_BOOTSTRAP_SERVERS')
emoji_input_topic = os.getenv('EMOJI_INPUT_TOPIC')
emoji_output_topic = os.getenv('EMOJI_OUTPUT_TOPIC')

CHECKPOINT_LOCATION = f"./tmp/event-window-output/{uuid.uuid4()}"
SPARK_VERSION = '3.5.2'
SCALA_VERSION = '2.12'
KAFKA_VERSION = '3.7.1'

packages = [
    f'org.apache.spark:spark-sql-kafka-0-10_{SCALA_VERSION}:{SPARK_VERSION}',
    'org.apache.kafka:kafka-clients:3.3.2'
]


os.environ['PYSPARK_PYTHON'] = sys.executable
os.environ['PYSPARK_DRIVER_PYTHON'] = sys.executable

# Set the Hadoop home directory
os.environ['HADOOP_HOME'] = 'D:\\programs\\winutils'

# Add Hadoop bin directory to PATH
os.environ['PATH'] += os.pathsep + os.path.join(os.environ['HADOOP_HOME'], 'bin')
os.environ['PYSPARK_SUBMIT_ARGS'] = f'--packages org.apache.spark:spark-streaming-kafka-0-10_{SCALA_VERSION}:{SPARK_VERSION},org.apache.spark:spark-sql-kafka-0-10_{SCALA_VERSION}:{SPARK_VERSION} pyspark-shell'

spark = SparkSession.builder \
    .master(app_mode) \
    .appName(app_name) \
    .config("spark.jars.packages",",".join(packages)) \
    .config("spark.sql.streaming.stateStore.stateSchemaCheck", "false") \
    .config("spark.sql.streaming.statefulOperator.checkCorrectness.enabled", "false") \
    .config("spark.sql.shuffle.partitions", "6") \
    .getOrCreate()

spark.sparkContext.setLogLevel("ERROR")
logger.info("=== Spark Session started =====")


# Function to print the values
def print_batch(batch_df, batch_id):
    logger.info(f"<==== Batch ID ====>: {batch_id}")
    batch_df.show(truncate=False)

# ============== SIMULATION =========================
# Simulate Kafka messages
schema = StructType([
    StructField("id", StringType(), True),
    StructField("Emoji", StringType(), True),
    StructField("Description", StringType(), True),
    StructField("Keywords", StringType(), True),
    StructField("Code", StringType(), True)
])

try:
    # read from Kafka Stream
    kafka_stream = spark \
        .readStream.format("kafka") \
        .option("kafka.bootstrap.servers", kafka_boostrap_server) \
        .option("subscribe", emoji_input_topic) \
        .option("startingOffsets", "latest") \
        .option("failOnDataLoss", "false") \
        .load()
    logging.info("===> Initial dataframe created successfully")
except Exception as e:
    logging.warning(f"Initial dataframe couldn't be created due to exception: {e}")

json_stream = kafka_stream.selectExpr("CAST(value AS STRING)") \
    .select(from_json("value", schema).alias("data")) \
    .select("data.*")

# Transform the data (e.g. add a timestamp column)
transformed_data = json_stream.withColumn("timestamp", current_timestamp())
transformed_data.printSchema()

# Add watermark and compute aggregates over a 2-second window

# Perform aggregation on ID
resultDF = transformed_data.withWatermark("timestamp", PROCEESSING_TIME) \
    .groupBy(window(col("timestamp"), PROCEESSING_TIME), "Emoji", "Code") \
    .agg(count("id").alias("count_of_id"))
# resultDF.printSchema()

# Collect all aggregated results into a single JSON array per window
aggregatedDF = resultDF.groupBy("window") \
    .agg(collect_list(struct("count_of_id", "Emoji", "Code")).alias("aggregated_data")) \
    .selectExpr("to_json(struct(window.start as window_start, window.end as window_end, aggregated_data)) as value")

# Write the results back to another Kafka topic
# kafka_stream = aggregatedDF.writeStream \
#     .outputMode("update") \
#     .option("truncate", False) \
#     .format("console") \
#     .option("checkpointLocation", "CHECKPOINT_LOCATION") \
#     .start()

# Write the data to Kafka
kafka_stream = aggregatedDF \
    .writeStream.format("kafka") \
    .option("kafka.bootstrap.servers", kafka_boostrap_server) \
    .option("topic", emoji_output_topic) \
    .option("checkpointLocation", CHECKPOINT_LOCATION) \
    .trigger(processingTime=PROCEESSING_TIME) \
    .start()

# Simplified write stream to console for validation

logger.info("----> Results to be written to Kafka")

kafka_stream.awaitTermination()

kafka_stream.stop()
