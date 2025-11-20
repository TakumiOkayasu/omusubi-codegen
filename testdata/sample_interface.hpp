// Sample interface for testing code generator
// This file represents a typical Omusubi framework interface

#ifndef SAMPLE_INTERFACE_HPP
#define SAMPLE_INTERFACE_HPP

#include <cstdint>

namespace omusubi {
namespace device {

/**
 * @brief Abstract interface for GPIO devices
 * @details This interface defines the contract for GPIO pin control
 */
class IGPIODevice {
public:
    /**
     * @brief Virtual destructor
     */
    virtual ~IGPIODevice() = default;

    /**
     * @brief Initialize the GPIO device
     * @return true if initialization succeeded, false otherwise
     */
    virtual bool initialize() = 0;

    /**
     * @brief Set the pin mode
     * @param pin Pin number
     * @param mode Pin mode (INPUT=0, OUTPUT=1)
     * @return true if successful, false otherwise
     */
    virtual bool setPinMode(uint8_t pin, uint8_t mode) = 0;

    /**
     * @brief Write digital value to pin
     * @param pin Pin number
     * @param value Digital value (LOW=0, HIGH=1)
     */
    virtual void digitalWrite(uint8_t pin, uint8_t value) = 0;

    /**
     * @brief Read digital value from pin
     * @param pin Pin number
     * @return Digital value (LOW=0, HIGH=1)
     */
    virtual uint8_t digitalRead(uint8_t pin) const = 0;

    /**
     * @brief Read analog value from pin
     * @param pin Pin number
     * @return Analog value (0-4095 for 12-bit ADC)
     */
    virtual uint16_t analogRead(uint8_t pin) const = 0;

protected:
    /**
     * @brief Validate pin number
     * @param pin Pin number to validate
     * @return true if valid, false otherwise
     */
    virtual bool validatePin(uint8_t pin) const = 0;
};

} // namespace device
} // namespace omusubi

#endif // SAMPLE_INTERFACE_HPP
