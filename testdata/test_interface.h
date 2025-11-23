// Test interface with .h extension
// This should generate .h + .cpp (not .c)

#ifndef TEST_INTERFACE_H
#define TEST_INTERFACE_H

#include <stdint.h>

namespace test {

/**
 * @brief Test interface for .h file handling
 */
class ITestDevice {
public:
    virtual ~ITestDevice() = default;

    /**
     * @brief Initialize the device
     * @return true if successful
     */
    virtual bool initialize() = 0;

    /**
     * @brief Get device status
     * @return Status code
     */
    virtual uint8_t getStatus() const = 0;
};

} // namespace test

#endif // TEST_INTERFACE_H
