-- +goose NO TRANSACTION
-- +goose Up
SELECT set_chunk_time_interval('ping_results', INTERVAL '1 day');
SELECT set_chunk_time_interval('tcp_results', INTERVAL '1 day');
SELECT set_chunk_time_interval('traceroute_results', INTERVAL '1 day');

SELECT set_chunk_time_interval('ping_rtt_sample_observations', INTERVAL '1 day');
SELECT set_chunk_time_interval('traceroute_hop_observations', INTERVAL '1 day');
SELECT set_chunk_time_interval('traceroute_edge_observations', INTERVAL '1 day');

-- +goose Down
SELECT set_chunk_time_interval('ping_results', INTERVAL '7 days');
SELECT set_chunk_time_interval('tcp_results', INTERVAL '7 days');
SELECT set_chunk_time_interval('traceroute_results', INTERVAL '7 days');

SELECT set_chunk_time_interval('ping_rtt_sample_observations', INTERVAL '7 days');
SELECT set_chunk_time_interval('traceroute_hop_observations', INTERVAL '7 days');
SELECT set_chunk_time_interval('traceroute_edge_observations', INTERVAL '7 days');
