package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/ymtdzzz/batch-tracing-sample/notification-manager/internal"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"

	amqp "github.com/rabbitmq/amqp091-go"
)

func initTracer() (*sdktrace.TracerProvider, error) {
	exporter, err := otlptracegrpc.New(
		context.Background(),
		otlptracegrpc.WithEndpoint("jaeger:4317"),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter, sdktrace.WithBatchTimeout(1*time.Second)),
		sdktrace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceNameKey.String("notification-batch"))),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tp, nil
}

func main() {
	tp, err := initTracer()
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			panic(err)
		}
	}()

	conn, err := amqp.Dial("amqp://guest:guest@my-queue:5672/")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		log.Printf("Sending messages. To exit press CTRL+C")
		for range c {
			close(c)
			log.Printf("Bye")
			os.Exit(0)
		}
	}()

	for {
		time.Sleep(1 * time.Second)

		var wg sync.WaitGroup

		rand.Seed(time.Now().UnixNano())
		waitNum := rand.Intn(3)
		for i := 0; i < waitNum; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				process(conn)
			}()
		}

		wg.Wait()
	}
}

func process(conn *amqp.Connection) {
	ctx, span := otel.Tracer("notification").Start(context.Background(), "produce")
	defer span.End()

	// In RabbitMQ, channel is not thread-safe.
	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"notification",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	msg, err := internal.NewRandomNotification().Encode()
	if err != nil {
		panic(err)
	}

	headers := amqp.NewConnectionProperties()
	carrier := internal.NewAMQPCarrier(headers)
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	err = ch.PublishWithContext(
		ctx,
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/octet-stream",
			Body:        msg,
			Headers:     headers,
		},
	)
	if err != nil {
		panic(err)
	}
	log.Println("Message has been sent")
}
