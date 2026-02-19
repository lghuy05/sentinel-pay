package com.example.alert_service.repository;

import java.util.List;
import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Lock;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;

import com.example.alert_service.entity.Transfer;
import com.example.alert_service.entity.TransferStatus;

import jakarta.persistence.LockModeType;

public interface TransferRepository extends JpaRepository<Transfer, Long> {
  @Lock(LockModeType.PESSIMISTIC_WRITE)
  @Query("""
        select t from Transfer t
        where t.transactionId = :transactionId
      """)
  Optional<Transfer> findByTransactionIdForUpdate(
      @Param("transactionId") String transactionId);

  List<Transfer> findTop100ByStatusOrderByUpdatedAtAsc(TransferStatus status);
}
