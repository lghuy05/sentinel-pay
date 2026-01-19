package com.example.blacklist_service.repository;

import java.util.Optional;
import java.util.List;

import org.springframework.data.jpa.repository.JpaRepository;

import com.example.blacklist_service.entity.BlacklistEntry;
import com.example.blacklist_service.entity.BlacklistType;

public interface BlacklistRepository extends JpaRepository<BlacklistEntry, Long> {
    Optional<BlacklistEntry> findByTypeAndValueAndActiveTrue(BlacklistType type, String value);
    List<BlacklistEntry> findAllByActiveTrue();
}
