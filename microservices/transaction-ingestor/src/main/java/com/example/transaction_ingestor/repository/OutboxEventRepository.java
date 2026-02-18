package com.example.transaction_ingestor.repository;

import java.time.Instant;
import java.util.List;

import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Lock;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;

import com.example.transaction_ingestor.entity.OutboxEvent;
import com.example.transaction_ingestor.entity.OutboxStatusType;

import jakarta.persistence.LockModeType;

public interface OutboxEventRepository
    extends JpaRepository<OutboxEvent, Long> {
  @Lock(LockModeType.PESSIMISTIC_WRITE)
  @Query("""
        select o from OutboxEvent o
        where o.status = :status and o.nextRetry
      <= :now
        order by o.createdAt asc
      """)
  List<OutboxEvent> findDue(
      @Param("status") OutboxStatusType status,
      @Param("now") Instant now,
      Pageable page);
}
