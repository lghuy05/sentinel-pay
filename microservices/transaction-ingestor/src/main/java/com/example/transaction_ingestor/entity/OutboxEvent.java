package com.example.transaction_ingestor.entity;

import java.time.Instant;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.EnumType;
import jakarta.persistence.Enumerated;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.Table;
import jakarta.persistence.UniqueConstraint;
import lombok.Getter;
import lombok.Setter;

@Getter
@Setter
@Entity
@Table(name = "outbox_event", uniqueConstraints = @UniqueConstraint(columnNames = { "aggregate_type", "aggregate_id",
    "event_type" }))

public class OutboxEvent {

  @Id
  @GeneratedValue(strategy = GenerationType.IDENTITY)
  private Long id;

  @Column(name = "aggregate_type", nullable = false)
  private String aggregateType;

  @Column(name = "aggregate_id", nullable = false)
  private String aggregateId;

  @Column(name = "event_type", nullable = false)
  private String eventType;

  @Column(name = "payload")
  private String payload;

  @Enumerated(EnumType.STRING)
  @Column(name = "status", nullable = false)
  private OutboxStatusType status;

  @Column(name = "attempt_count", nullable = false)
  private int attemptCount;

  @Column(name = "created_at", nullable = false)
  private Instant createdAt;

  @Column(name = "next_retry_at")
  private Instant nextRetry;

  @Column(name = "published_at")
  private Instant publishedAt;

  @Column(name = "last_error")
  private String lastError;

}
