CREATE CONTINUOUS QUERY "downsampled_bus_data_mean" ON transportation BEGIN SELECT mean(value) as value INTO "autogen.bus_data_mean" FROM bus_data GROUP BY time(1m) END
 