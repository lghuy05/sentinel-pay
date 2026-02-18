package com.example.transaction_ingestor;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.scheduling.annotation.EnableScheduling;

@SpringBootApplication
@EnableScheduling
public class TransactionIngestorApplication {

  public static void main(String[] args) {
    SpringApplication.run(TransactionIngestorApplication.class, args);
  }

}
